// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/juju/clock"
	"github.com/juju/collections/set"
	"github.com/juju/errors"
	"github.com/juju/loggo"
	"github.com/juju/mgo/v2"
	"github.com/juju/mgo/v2/bson"
	"github.com/juju/retry"
	"github.com/juju/utils/v3"
	"github.com/juju/version/v2"
	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

var (
	// MgoServer is a shared mongo server used by tests.
	MgoServer = &MgoInstance{}
	logger    = loggo.GetLogger("juju.testing")

	// regular expression to match output of mongod
	waitingForConnectionsRe = regexp.MustCompile(".*[W|w]aiting for connections.*")

	mongo32 = version.Number{Major: 3, Minor: 2}

	// After version 3.2 we shouldn't use --nojournal - it makes the
	// WiredTiger storage engine much slower.
	// https://jira.mongodb.org/browse/SERVER-21198
	useJournalMongoVersion = mongo32

	// From mongo 3.2 onwards, we can specify a storage engine.
	storageEngineMongoVersion = mongo32

	installedMongod mongodCache
)

const (
	// Maximum number of times to attempt starting mongod.
	maxStartMongodAttempts = 5

	// The default password to use when connecting to the mongo database.
	DefaultMongoPassword = "conn-from-name-secret"
)

// Certs holds the certificates and keys required to make a secure
// SSL connection.
type Certs struct {
	// CACert holds the CA certificate. This must certify the private key that
	// was used to sign the server certificate.
	CACert *x509.Certificate
	// ServerCert holds the certificate that certifies the server's
	// private key.
	ServerCert *x509.Certificate
	// ServerKey holds the server's private key.
	ServerKey *rsa.PrivateKey
}

type MgoInstance struct {
	// addr holds the address of the MongoDB server
	addr string

	// MgoPort holds the port of the MongoDB server.
	port int

	// server holds the running MongoDB command.
	server *exec.Cmd

	// exited receives a value when the mongodb server exits.
	exited <-chan struct{}

	// dir holds the directory that MongoDB is running in.
	dir string

	// certs holds certificates for the TLS connection.
	certs *Certs

	// Params is a list of additional parameters that will be passed to
	// the mongod application
	Params []string

	// EnableReplicaSet will pass the right parameters to --replSet and call
	// replSetInitiate when appropriate.
	EnableReplicaSet bool

	// EnableAuth enables authentication/authorization.
	EnableAuth bool

	// WithoutV8 is true if we believe this Mongo doesn't actually have the
	// V8 engine
	WithoutV8 bool

	// MaxTransactionLockRequestTimeout is used for the mongo
	// maxTransactionLockRequestTimeoutMillis server setting (v4+).
	MaxTransactionLockRequestTimeout time.Duration
}

// Addr returns the address of the MongoDB server.
func (m *MgoInstance) Addr() string {
	return m.addr
}

// Port returns the port of the MongoDB server.
func (m *MgoInstance) Port() int {
	return m.port
}

// SSLEnabled reports whether or not SSL is enabled for the MongoDB server.
func (m *MgoInstance) SSLEnabled() bool {
	return m.certs != nil
}

// We specify a timeout to mgo.Dial, to prevent
// mongod failures hanging the tests.
const mgoDialTimeout = 60 * time.Second

// MgoSuite is a suite that deletes all content from the shared MongoDB
// server at the end of every test and supplies a connection to the shared
// MongoDB server.
type MgoSuite struct {
	Session *mgo.Session

	// DebugMgo controls whether SetUpSuite enables mgo logging and
	// debugging. Set this before calling SetUpSuite. Enabling either
	// logging or debugging in mgo adds a significant overhead to the
	// Juju tests, so they are disabled by default.
	DebugMgo bool

	// SkipTestCleanup controls collection cleanup in TearDownTest.
	// When set to true, TearDownTest will not delete collections.
	SkipTestCleanup bool
}

// generatePEM receives server certificate and the server private key
// and creates a PEM file in the given path.
func generatePEM(path string, serverCert *x509.Certificate, serverKey *rsa.PrivateKey) error {
	pemFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open %q for writing: %v", path, err)
	}
	defer pemFile.Close()
	err = pem.Encode(pemFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCert.Raw,
	})
	if err != nil {
		return fmt.Errorf("failed to write cert to %q: %v", path, err)
	}
	err = pem.Encode(pemFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})
	if err != nil {
		return fmt.Errorf("failed to write private key to %q: %v", path, err)
	}
	return nil
}

// getHome for robust detection of HOME directory for use on Linux with
// snaps.
func getHome() (string, error) {
	targetUID := strconv.Itoa(os.Getuid())
	passwd, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return "", errors.Trace(err)
	}
	lines := strings.Split(string(passwd), "\n")
	for _, line := range lines {
		passwdEntry := strings.Split(line, ":")
		if len(passwdEntry) != 7 {
			// Invalid passwd entry.
			continue
		}
		uidEntry := passwdEntry[2]
		if uidEntry != targetUID {
			// Not the user we are looking for.
			continue
		}
		home := passwdEntry[5]
		return home, nil
	}
	return "", errors.NotFoundf("UNIX user %s", targetUID)
}

// Start starts a MongoDB server in a temporary directory.
func (inst *MgoInstance) Start(certs *Certs) error {
	var err error
	mongopath, vers, err := installedMongod.Get()
	if err != nil {
		return err
	}

	dbdir := ""

	// Check for snap confined mongod.
	if runtime.GOOS == "linux" && mongopath == "/snap/bin/juju-db.mongod" {
		home, err := getHome()
		if err != nil {
			return errors.Annotatef(err, "failed to find HOME directory")
		}
		base := path.Join(home, "snap/juju-db/current/tmp")
		err = os.Mkdir(base, 0755)
		if os.IsExist(err) {
			// do nothing
		} else if err != nil {
			return err
		}
		dbdir, err = ioutil.TempDir(base, "test-mgo")
		if err != nil {
			return errors.Annotatef(err, "failed to make directory for confined juju-db snap")
		}
	} else {
		dbdir, err = ioutil.TempDir("", "test-mgo")
		if err != nil {
			return err
		}
	}

	logger.Debugf("starting mongo in %s", dbdir)

	// Give them all the same keyfile so they can talk appropriately.
	keyFilePath := filepath.Join(dbdir, "keyfile")
	err = ioutil.WriteFile(keyFilePath, []byte("not very secret"), 0600)
	if err != nil {
		return fmt.Errorf("cannot write key file: %v", err)
	}

	if certs != nil {
		// Generate and save the server.pem file.
		pemPath := filepath.Join(dbdir, "server.pem")
		if err = generatePEM(pemPath, certs.ServerCert, certs.ServerKey); err != nil {
			return fmt.Errorf("cannot write cert/key PEM: %v", err)
		}
		inst.certs = certs
	}

	// Attempt to start mongo up to maxStartMongodAttempts times,
	// as the port we choose may be taken from us in the mean time.
	for i := 0; i < maxStartMongodAttempts; i++ {
		inst.port = FindTCPPort()
		inst.addr = fmt.Sprintf("localhost:%d", inst.port)
		inst.dir = dbdir
		err = inst.run(vers)
		switch err.(type) {
		case addrAlreadyInUseError:
			logger.Debugf("failed to start mongo: %v, trying another port", err)
			continue
		case nil:
			logger.Debugf("started mongod pid %d in %s on port %d", inst.server.Process.Pid, dbdir, inst.port)
		default:
			inst.addr = ""
			inst.port = 0
			os.RemoveAll(inst.dir)
			inst.dir = ""
			logger.Warningf("failed to start mongo: %v", err)
		}
		break
	}
	return err
}

// run runs the MongoDB server at the
// address and directory already configured.
func (inst *MgoInstance) run(vers version.Number) error {
	if inst.server != nil {
		panic("mongo server is already running")
	}

	mgoport := strconv.Itoa(inst.port)
	mgoargs := []string{
		"--dbpath", inst.dir,
		"--port", mgoport,
		"--oplogSize", "10",
		"--ipv6",
		"--setParameter", "enableTestCommands=1",
		// You can set this if you want to see all queries that are
		// being run against Mongodb. We don't enable it by default
		// because it ends up being very chatty.
		// "--setParameter", "logComponentVerbosity={verbosity:1}",
	}
	if runtime.GOOS != "windows" {
		mgoargs = append(mgoargs, "--nounixsocket")
	}
	if inst.EnableAuth {
		mgoargs = append(mgoargs,
			"--auth",
			"--keyFile", filepath.Join(inst.dir, "keyfile"),
		)
	}
	if inst.EnableReplicaSet {
		mgoargs = append(mgoargs, "--replSet=juju")
	}
	if vers.Major >= 4 && inst.MaxTransactionLockRequestTimeout > 0 {
		arg := fmt.Sprintf("maxTransactionLockRequestTimeoutMillis=%d",
			inst.MaxTransactionLockRequestTimeout/time.Millisecond)
		mgoargs = append(mgoargs, "--setParameter", arg)
	}
	if inst.certs != nil {
		mgoargs = append(mgoargs,
			"--sslMode", "requireSSL",
			"--sslPEMKeyFile", filepath.Join(inst.dir, "server.pem"),
			"--sslPEMKeyPassword=ignored")
	}

	mongopath, version, err := installedMongod.Get()
	if err != nil {
		return err
	}
	logger.Debugf("using mongod at: %q (version=%s)", mongopath, version)

	if version.Compare(useJournalMongoVersion) == -1 {
		mgoargs = append(mgoargs, "--nojournal")
	}
	if version.Compare(storageEngineMongoVersion) >= 0 {
		storageEngine := mongoStorageEngine()
		if storageEngine != "" {
			mgoargs = append(mgoargs, "--storageEngine", storageEngine)
			if storageEngine == "mmapv1" {
				mgoargs = append(mgoargs, "--nssize", "1",
					"--noprealloc",
					"--smallfiles")
			}
		}
	}

	if inst.Params != nil {
		mgoargs = append(mgoargs, inst.Params...)
	}
	if mongopath == "/usr/lib/juju/bin/mongod" || mongopath == "/usr/lib/juju/mongo3.2/bin/mongod" {
		inst.WithoutV8 = true
	}
	server := exec.Command(mongopath, mgoargs...)
	out, err := server.StdoutPipe()
	if err != nil {
		return err
	}
	server.Stderr = server.Stdout
	exited := make(chan struct{})
	started := make(chan error)
	listening := make(chan error, 1)
	go func() {
		err := <-started
		if err != nil {
			close(listening)
			close(exited)
			return
		}
		// Wait until the server is listening.
		var buf bytes.Buffer
		prefix := fmt.Sprintf("mongod:%v", mgoport)
		if readUntilMatching(prefix, io.TeeReader(out, &buf), waitingForConnectionsRe) {
			listening <- nil
		} else {
			err := fmt.Errorf("mongod failed to listen on port %v", mgoport)
			if strings.Contains(buf.String(), "already in use") {
				err = addrAlreadyInUseError{err}
			}
			listening <- err
		}
		// Capture the last 100 lines of output from mongod, to log
		// in the event of unclean exit.
		lines := readLastLines(prefix, io.MultiReader(&buf, out), 100)
		err = server.Wait()
		exitErr, _ := err.(*exec.ExitError)
		if err == nil || exitErr != nil && exitErr.Exited() {
			// mongodb has exited without being killed, so print the
			// last few lines of its log output.
			logger.Errorf("mongodb has exited without being killed")
			for _, line := range lines {
				logger.Errorf("mongod: %s", line)
			}
		}
		close(exited)
	}()
	inst.exited = exited
	err = server.Start()
	started <- err
	if err != nil {
		return err
	}
	err = <-listening
	close(listening)
	if err != nil {
		return err
	}
	inst.server = server
	if inst.EnableReplicaSet {
		session := inst.MustDialDirect()
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		var res bson.M
		if err := session.Run(bson.D{{"replSetInitiate", nil}}, &res); err != nil {
			return err
		}
		logger.Debugf("mongodb initializing replicaset returned: %v", res)
	}
	return nil
}

func mongoStorageEngine() string {
	storageEngine := os.Getenv("JUJU_MONGO_STORAGE_ENGINE")
	if storageEngine != "" {
		return storageEngine
	}
	switch runtime.GOARCH {
	case "amd64":
		// Wired tiger is always available from 3.x onwards.
		return "wiredTiger"
	}
	return "" // use the default
}

// mongodCache looks up mongod path and version and caches the result.
type mongodCache struct {
	sync.Mutex
	path    string
	version version.Number
	done    bool
}

func MongodVersion() (version.Number, error) {
	_, v, err := installedMongod.Get()
	return v, err
}

func (cache *mongodCache) Get() (string, version.Number, error) {
	cache.Lock()
	defer cache.Unlock()
	if !cache.done {
		var err error
		cache.path, err = getMongod()
		if err != nil {
			return "", version.Zero, errors.Trace(err)
		}
		cache.version, err = detectMongoVersion(cache.path)
		if err != nil {
			return "", version.Zero, errors.Trace(err)
		}
		cache.done = true
	}
	return cache.path, cache.version, nil
}

func getMongod() (string, error) {
	// Prefer $JUJU_MONGOD and then newer MongoDBs.
	var paths []string
	if path := os.Getenv("JUJU_MONGOD"); path != "" {
		paths = append(paths, path)
	}

	if runtime.GOOS == "linux" {
		// Snaps are only supported on linux for now.
		paths = append(paths, "/snap/bin/juju-db.mongod")
	}

	paths = append(paths,
		"/usr/lib/juju/mongo3.2/bin/mongod",
		"mongod",
		"/usr/lib/juju/bin/mongod",
		"/usr/bin/mongod",       // bionic
		"/usr/local/bin/mongod", // Needed on CentOS where $PATH is being completely removed
	)
	var err error
	var mongopath string
	for _, path := range paths {
		mongopath, err = exec.LookPath(path)
		if err == nil {
			return mongopath, nil
		}
		logger.Debugf("failed to find %q: %v", path, err)
	}
	return "", err
}

// The mongod --version line starts with this prefix.
const versionLinePrefix = "db version v"

func detectMongoVersion(mongoPath string) (version.Number, error) {
	output, err := exec.Command(mongoPath, "--version").Output()
	if err != nil {
		return version.Zero, errors.Trace(err)
	}
	// Read the first line of the output with a scanner (to handle
	// newlines in a cross-platform way).
	scanner := bufio.NewScanner(bytes.NewReader(output))
	versionLine := ""
	if scanner.Scan() {
		versionLine = scanner.Text()
	}
	if scanner.Err() != nil {
		return version.Zero, errors.Trace(scanner.Err())
	}
	if !strings.HasPrefix(versionLine, versionLinePrefix) {
		return version.Zero, errors.New("couldn't get mongod version - no version line")
	}
	ver, err := version.Parse(versionLine[len(versionLinePrefix):])
	if err != nil {
		return version.Zero, errors.Trace(err)
	}
	return ver, nil
}

func (inst *MgoInstance) kill(sig os.Signal) {
	inst.server.Process.Signal(sig)
	<-inst.exited
	inst.server = nil
	inst.exited = nil
}

func (inst *MgoInstance) killAndCleanup(sig os.Signal) {
	if inst.server != nil {
		logger.Debugf("killing mongod pid %d in %s on port %d with %s", inst.server.Process.Pid, inst.dir, inst.port, sig)
		inst.kill(sig)
		os.RemoveAll(inst.dir)
		inst.addr, inst.dir = "", ""
	}
}

// Destroy kills mongod and cleans up its data directory.
func (inst *MgoInstance) Destroy() {
	inst.killAndCleanup(os.Kill)
}

// Restart restarts the mongo server, useful for
// testing what happens when a state server goes down.
func (inst *MgoInstance) Restart() {
	logger.Debugf("restarting mongod pid %d in %s on port %d", inst.server.Process.Pid, inst.dir, inst.port)
	inst.kill(os.Kill)
	if err := inst.Start(inst.certs); err != nil {
		panic(err)
	}
}

// MgoTestPackage should be called to register the tests for any package
// that requires a MongoDB server. If certs is non-nil, a secure SSL connection
// will be used from client to server.
func MgoTestPackage(t *testing.T, certs *Certs) {
	if err := MgoServer.Start(certs); err != nil {
		t.Fatal(err)
	}
	defer MgoServer.Destroy()
	gc.TestingT(t)
}

type mgoLogger struct {
	logger loggo.Logger
}

// Output implements the mgo log_Logger interface.
func (s *mgoLogger) Output(calldepth int, message string) error {
	s.logger.LogCallf(calldepth, loggo.TRACE, message)
	return nil
}

func (s *MgoSuite) SetUpSuite(c *gc.C) {
	if s.DebugMgo {
		mgo.SetLogger(&mgoLogger{loggo.GetLogger("mgo")})
		mgo.SetDebug(true)
	}
	if MgoServer.addr == "" {
		c.Fatalf("No Mongo Server Address, MgoSuite tests must be run with MgoTestPackage")
	}
	mgo.SetStats(true)
	// Make tests that use password authentication faster.
	utils.FastInsecureHash = true
	mgo.ResetStats()
	session, err := MgoServer.Dial()
	c.Assert(err, jc.ErrorIsNil)
	defer session.Close()
	err = dropAll(session)
	c.Assert(err, jc.ErrorIsNil)
}

// readUntilMatching reads lines from the given reader until the reader
// is depleted or a line matches the given regular expression.
func readUntilMatching(prefix string, r io.Reader, re *regexp.Regexp) bool {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		logger.Tracef("%s: %s", prefix, line)
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// readLastLines reads lines from the given reader and returns
// the last n non-empty lines, ignoring empty lines.
func readLastLines(prefix string, r io.Reader, n int) []string {
	sc := bufio.NewScanner(r)
	lines := make([]string, n)
	i := 0
	for sc.Scan() {
		if line := strings.TrimRight(sc.Text(), "\n"); line != "" {
			logger.Tracef("%s: %s", prefix, line)
			lines[i%n] = line
			i++
		}
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	final := make([]string, 0, n+1)
	if i > n {
		final = append(final, fmt.Sprintf("[%d lines omitted]", i-n))
	}
	for j := 0; j < n; j++ {
		if line := lines[(j+i)%n]; line != "" {
			final = append(final, line)
		}
	}
	return final
}

func (s *MgoSuite) TearDownSuite(c *gc.C) {
	err := MgoServer.Reset()
	c.Assert(err, jc.ErrorIsNil)
	utils.FastInsecureHash = false
	if s.DebugMgo {
		mgo.SetDebug(false)
		mgo.SetLogger(nil)
	}
}

// MustDial returns a new connection to the MongoDB server, and panics on
// errors.
func (inst *MgoInstance) MustDial() *mgo.Session {
	s, err := mgo.DialWithInfo(inst.DialInfo())
	if err != nil {
		panic(err)
	}
	return s
}

// Dial returns a new connection to the MongoDB server.
func (inst *MgoInstance) Dial() (*mgo.Session, error) {
	var session *mgo.Session
	err := retry.Call(retry.CallArgs{
		Func: func() error {
			var err error
			session, err = mgo.DialWithInfo(inst.DialInfo())
			return err
		},
		// Only interested in retrying the intermittent
		// 'unexpected message'.
		IsFatalError: func(err error) bool {
			return !strings.HasSuffix(err.Error(), "unexpected message")
		},
		Delay:    time.Millisecond,
		Clock:    clock.WallClock,
		Attempts: 5,
	})
	return session, err
}

// DialInfo returns information suitable for dialling the
// receiving MongoDB instance.
func (inst *MgoInstance) DialInfo() *mgo.DialInfo {
	return MgoDialInfo(inst.certs, inst.addr)
}

// DialDirect returns a new direct connection to the shared MongoDB server. This
// must be used if you're connecting to a replicaset that hasn't been initiated
// yet.
func (inst *MgoInstance) DialDirect() (*mgo.Session, error) {
	info := inst.DialInfo()
	info.Direct = true
	return mgo.DialWithInfo(info)
}

// MustDialDirect works like DialDirect, but panics on errors.
func (inst *MgoInstance) MustDialDirect() *mgo.Session {
	session, err := inst.DialDirect()
	if err != nil {
		panic(err)
	}
	return session
}

// MgoDialInfo returns a DialInfo suitable
// for dialling an MgoInstance at any of the
// given addresses, optionally using TLS.
func MgoDialInfo(certs *Certs, addrs ...string) *mgo.DialInfo {
	var dial func(addr net.Addr) (net.Conn, error)
	if certs != nil {
		pool := x509.NewCertPool()
		pool.AddCert(certs.CACert)
		tlsConfig := &tls.Config{
			RootCAs:    pool,
			ServerName: "anything",
		}
		dial = func(addr net.Addr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			if err != nil {
				logger.Debugf("tls.Dial(%s) failed with %v", addr, err)
				return nil, err
			}
			return conn, nil
		}
	} else {
		dial = func(addr net.Addr) (net.Conn, error) {
			conn, err := net.Dial("tcp", addr.String())
			if err != nil {
				logger.Debugf("net.Dial(%s) failed with %v", addr, err)
				return nil, err
			}
			return conn, nil
		}
	}
	return &mgo.DialInfo{Addrs: addrs, Dial: dial, Timeout: mgoDialTimeout}
}

func clearDatabases(session *mgo.Session) error {
	databases, err := session.DatabaseNames()
	if err != nil {
		return errors.Annotate(err, "failed to list database names")
	}
	for _, name := range databases {
		if name == "local" || name == "config" {
			// local has lots of things like oplog.rs or
			// replset.*
			// config contains things like config.transactions
			// none of those are safe to delete.
			continue
		}
		err = clearCollections(session.DB(name))
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func clearCollections(db *mgo.Database) error {
	capped, err := listCappedCollections(db)
	if err != nil {
		return errors.Annotatef(err, "getting capped collection list")
	}
	collectionNames, err := db.CollectionNames()
	if err != nil {
		return errors.Trace(err)
	}

	// The "logs" db is special because tests create namespaced capped
	// collections for their logs. Contrary to the other DBs where we
	// just flush their contents, we need to drop any collection in the
	// "logs" db to avoid an unbounded growth of capped collections.
	if db.Name == "logs" {
		for _, name := range collectionNames {
			if err = db.C(name).DropCollection(); err != nil {
				return errors.Trace(err)
			}
		}
		return nil
	}

	for _, name := range collectionNames {
		if strings.HasPrefix(name, "system.") {
			continue
		}
		collection := db.C(name)
		clearFunc := clearNormalCollection
		if capped.Contains(name) {
			clearFunc = clearCappedCollection
		}
		err = clearFunc(collection)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func listCappedCollections(db *mgo.Database) (set.Strings, error) {
	// Mostly pulled from mgo.DB.ListCollections
	names := set.NewStrings()
	var result struct {
		Collections []bson.Raw
		Cursor      struct {
			FirstBatch []bson.Raw `bson:"firstBatch"`
			NextBatch  []bson.Raw `bson:"nextBatch"`
			NS         string     `bson:"ns"`
			ID         int64      `bson:"id"`
		}
	}
	err := db.Run(bson.D{{"listCollections", 1}, {"cursor", bson.D{{"batchSize", 10}}}}, &result)
	if err != nil {
		return nil, errors.Trace(err)
	}
	firstBatch := result.Collections
	if firstBatch == nil {
		firstBatch = result.Cursor.FirstBatch
	}
	var iter *mgo.Iter
	ns := strings.SplitN(result.Cursor.NS, ".", 2)
	if len(ns) < 2 {
		iter = db.C("").NewIter(nil, firstBatch, result.Cursor.ID, nil)
	} else {
		iter = db.Session.DB(ns[0]).C(ns[1]).NewIter(nil, firstBatch, result.Cursor.ID, nil)
	}
	var coll struct {
		Name    string `bson:"name"`
		Options struct {
			Capped bool `bson:"capped"`
		} `bson:"options"`
	}
	for iter.Next(&coll) {
		if coll.Options.Capped {
			names.Add(coll.Name)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Trace(err)
	}
	return names, nil
}

func clearNormalCollection(collection *mgo.Collection) error {
	_, err := collection.RemoveAll(bson.M{})
	return err
}

func clearCappedCollection(collection *mgo.Collection) error {
	// This is a test command - relies on the enableTestCommands
	// setting being passed to mongo at startup.
	return collection.Database.Run(bson.D{{"emptycapped", collection.Name}}, nil)
}

func (s *MgoSuite) SetUpTest(c *gc.C) {
	s.Session = nil
	mgo.ResetStats()
	session, err := MgoServer.Dial()
	c.Assert(err, jc.ErrorIsNil)
	s.Session = session
}

// Reset deletes all content from the MongoDB server.
func (inst *MgoInstance) Reset() error {
	err := inst.EnsureRunning()
	if err != nil {
		return errors.Trace(err)
	}
	session, err := inst.Dial()
	if err != nil {
		return errors.Annotate(err, "inst.Dial() failed")
	}
	defer session.Close()

	dbnames, ok, err := resetAdminPasswordAndFetchDBNames(session)
	if err != nil {
		return errors.Trace(err)
	}
	if !ok {
		// We restart it to regain access.  This should only
		// happen when tests fail.
		logger.Infof("restarting MongoDB server after unauthorized access")
		inst.Destroy()
		err := inst.Start(inst.certs)
		return errors.Annotatef(err, "inst.Start(%v) failed", inst.certs)
	}
	logger.Infof("reset successfully reset admin password")
	for _, name := range dbnames {
		switch name {
		case "local", "config", "admin":
			// don't delete these
			continue
		}
		if err := session.DB(name).DropDatabase(); err != nil {
			return errors.Annotatef(err, "cannot drop MongoDB database %v", name)
		}
	}
	return nil
}

// dropAll drops all databases apart from admin, local and config.
func dropAll(session *mgo.Session) (err error) {
	names, err := session.DatabaseNames()
	if err != nil {
		return err
	}
	for _, name := range names {
		switch name {
		case "admin", "local", "config":
		default:
			err = session.DB(name).DropDatabase()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// resetAdminPasswordAndFetchDBNames logs into the database with a
// plausible password and returns all the database's db names. We need
// to try several passwords because we don't know what state the mongo
// server is in when Reset is called. If the test has set a custom
// password, we're out of luck, but if they are using
// DefaultStatePassword, we can succeed.
func resetAdminPasswordAndFetchDBNames(session *mgo.Session) ([]string, bool, error) {
	// First try with no password
	dbnames, err := session.DatabaseNames()
	if err == nil {
		return dbnames, true, nil
	}
	if !isUnauthorized(err) {
		return nil, false, errors.Trace(err)
	}
	// Then try the two most likely passwords in turn.
	for _, password := range []string{
		DefaultMongoPassword,
		utils.UserPasswordHash(DefaultMongoPassword, utils.CompatSalt),
	} {
		admin := session.DB("admin")
		if err := admin.Login("admin", password); err != nil {
			logger.Errorf("failed to log in with password %q", password)
			continue
		}
		dbnames, err := session.DatabaseNames()
		if err == nil {
			if err := admin.RemoveUser("admin"); err != nil {
				return nil, false, errors.Trace(err)
			}
			return dbnames, true, nil
		}
		if !isUnauthorized(err) {
			return nil, false, errors.Trace(err)
		}
		logger.Infof("unauthorized access when getting database names; password %q", password)
	}
	return nil, false, errors.Trace(err)
}

// isUnauthorized is a copy of the same function in state/open.go.
func isUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	// Some unauthorized access errors have no error code,
	// just a simple error string.
	if err.Error() == "auth fails" {
		return true
	}
	if err, ok := err.(*mgo.QueryError); ok {
		return err.Code == 10057 ||
			err.Message == "need to login" ||
			err.Message == "unauthorized"
	}
	return false
}

func (inst *MgoInstance) EnsureRunning() error {
	// If the server has already been destroyed for testing purposes,
	// just start it again.
	if inst.Addr() == "" {
		logger.Debugf("restarting mongo instance")
		err := inst.Start(inst.certs)
		return errors.Annotatef(err, "inst.Start(%v) failed", inst.certs)
	}
	return nil
}

func (s *MgoSuite) TearDownTest(c *gc.C) {
	if s.Session == nil {
		c.Fatal("SetUpTest failed")
	}

	err := MgoServer.EnsureRunning()
	c.Assert(err, jc.ErrorIsNil)

	// If the Session we have doesn't know about
	// the address of the server, then we should reconnect.
	foundAddress := false
	for _, addr := range s.Session.LiveServers() {
		if addr == MgoServer.Addr() {
			foundAddress = true
			break
		}
	}

	if !foundAddress {
		// The test has killed the server - reconnect.
		s.Session.Close()
		s.Session, err = MgoServer.Dial()
		c.Assert(err, jc.ErrorIsNil)
	}

	if !s.SkipTestCleanup {
		// Rather than dropping the databases (which is very slow in Mongo
		// 3.2) we clear all of the collections.
		err = clearDatabases(s.Session)
		c.Assert(err, jc.ErrorIsNil)
	}
	s.Session.Close()
	s.Session = nil

	for i := 0; ; i++ {
		stats := mgo.GetStats()
		if stats.SocketsInUse == 0 && stats.SocketsAlive == 0 {
			break
		}
		if i == 20 {
			c.Fatal("Test left sockets in a dirty state")
		}
		c.Logf("Waiting for sockets to die: %d in use, %d alive", stats.SocketsInUse, stats.SocketsAlive)
		time.Sleep(500 * time.Millisecond)
	}
}

// ProxiedSession represents a mongo session that's
// proxied through a TCPProxy instance.
type ProxiedSession struct {
	*mgo.Session
	*TCPProxy
}

// NewProxiedSession returns a ProxiedSession instance that holds a
// mgo.Session that directs through a TCPProxy instance to the testing
// mongoDB server, and the proxy instance itself. This allows tests to
// check what happens when mongo connections are broken.
//
// The returned value should be closed after use.
func NewProxiedSession(c *gc.C) *ProxiedSession {
	mgoInfo := MgoServer.DialInfo()
	c.Assert(mgoInfo.Addrs, gc.HasLen, 1)
	proxy := NewTCPProxy(c, mgoInfo.Addrs[0])
	mgoInfo.Addrs = []string{proxy.Addr()}
	session, err := mgo.DialWithInfo(mgoInfo)
	c.Assert(err, gc.IsNil)
	err = session.Ping()
	c.Assert(err, jc.ErrorIsNil)
	return &ProxiedSession{
		Session:  session,
		TCPProxy: proxy,
	}
}

// Close closes s.Session and s.TCPProxy.
func (s *ProxiedSession) Close() {
	s.Session.Close()
	s.TCPProxy.Close()
}

// FindTCPPort finds an unused TCP port and returns it.
// Use of this function has an inherent race condition - another
// process may claim the port before we try to use it.
// We hope that the probability is small enough during
// testing to be negligible.
func FindTCPPort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

type addrAlreadyInUseError struct {
	error
}

// IsolatedMgoSuite is a convenience type that combines the functionality
// IsolationSuite and MgoSuite.
type IsolatedMgoSuite struct {
	IsolationSuite
	MgoSuite
}

func (s *IsolatedMgoSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.MgoSuite.SetUpSuite(c)
}

func (s *IsolatedMgoSuite) TearDownSuite(c *gc.C) {
	s.MgoSuite.TearDownSuite(c)
	s.IsolationSuite.TearDownSuite(c)
}

func (s *IsolatedMgoSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.MgoSuite.SetUpTest(c)
}

func (s *IsolatedMgoSuite) TearDownTest(c *gc.C) {
	s.MgoSuite.TearDownTest(c)
	s.IsolationSuite.TearDownTest(c)
}
