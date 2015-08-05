package libkbfs

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"

	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/protocol/go"
	"golang.org/x/net/context"
)

func getMDServerAddr() string {
	// XXX TODO: the source of this will likely change soon
	return os.Getenv(EnvMDServerAddr)
}

func useLocalMDServer() bool {
	return len(getMDServerAddr()) == 0
}

func useLocalKeyServer() bool {
	// currently the remote MD server also acts as the key server.
	return useLocalMDServer()
}

func makeMDServer(config Config, serverRootDir *string) (
	MDServer, error) {
	var err error
	var mdServer MDServer
	if useLocalMDServer() {
		if serverRootDir == nil {
			// local in-memory MD server
			mdServer, err = NewMDServerMemory(config)
			if err != nil {
				return nil, err
			}
		}
		// local persistent MD server
		handlePath := filepath.Join(*serverRootDir, "kbfs_handles")
		mdPath := filepath.Join(*serverRootDir, "kbfs_md")
		revPath := filepath.Join(*serverRootDir, "kbfs_revisions")
		mdServer, err = NewMDServerLocal(
			config, handlePath, mdPath, revPath)
		if err != nil {
			return nil, err
		}
	} else {
		// this can't fail. reconnection attempts will be automatic.
		mdServer = NewMDServerRemote(context.TODO(), config, getMDServerAddr())
	}
	return mdServer, nil
}

func makeKeyServer(config Config, serverRootDir *string) (
	KeyServer, error) {
	var err error
	var keyServer KeyServer
	if useLocalKeyServer() {
		if serverRootDir == nil {
			// local in-memory key server
			keyServer, err = NewKeyServerMemory(config)
			if err != nil {
				return nil, err
			}
		}
		// local persistent key server
		keyPath := filepath.Join(*serverRootDir, "kbfs_key")
		keyServer, err = NewKeyServerLocal(config, keyPath)
		if err != nil {
			return nil, err
		}
	} else {
		// currently the remote MD server also acts as the key server.
		keyServer = config.MDServer().(*MDServerRemote)
	}
	return keyServer, nil
}

func makeBlockServer(config Config, serverRootDir *string) (BlockServer, error) {
	if serverRootDir == nil {
		return NewBlockServerMemory(config)
	}

	blockPath := filepath.Join(*serverRootDir, "kbfs_block")
	return NewBlockServerLocal(config, blockPath)
}

// Init initializes a config and returns it. If localUser is
// non-empty, libkbfs does not communicate to any remote servers and
// instead uses fake implementations of various servers.
//
// If serverRootDir is nil, an in-memory server is used. If it is
// non-nil and points to the empty string, the current working
// directory is used. Otherwise, the pointed-to string is treated as a
// path.
//
// onInterruptFn is called whenever an interrupt signal is received
// (e.g., if the user hits Ctrl-C).
//
// Init should be called at the beginning of main. Shutdown (see
// below) should then be called at the end of main (usually via
// defer).
func Init(localUser string, serverRootDir *string, cpuProfilePath, memProfilePath string, onInterruptFn func()) (Config, error) {
	if cpuProfilePath != "" {
		// Let the GC/OS clean up the file handle.
		f, err := os.Create(cpuProfilePath)
		if err != nil {
			return nil, err
		}
		pprof.StartCPUProfile(f)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		_ = <-sigchan

		Shutdown(memProfilePath)

		if onInterruptFn != nil {
			onInterruptFn()
		}

		os.Exit(1)
	}()

	config := NewConfigLocal()

	mdServer, err := makeMDServer(config, serverRootDir)
	if err != nil {
		return nil, fmt.Errorf("problem creating MD server: %v", err)
	}
	config.SetMDServer(mdServer)

	keyServer, err := makeKeyServer(config, serverRootDir)
	if err != nil {
		return nil, fmt.Errorf("problem creating key server: %v", err)
	}
	config.SetKeyServer(keyServer)

	bserv, err := makeBlockServer(config, serverRootDir)
	if err != nil {
		return nil, fmt.Errorf("cannot open block database: %v", err)
	}
	config.SetBlockServer(bserv)

	libkb.G.Init()
	libkb.G.ConfigureConfig()
	libkb.G.ConfigureLogging()
	libkb.G.ConfigureCaches()
	libkb.G.ConfigureMerkleClient()

	client.InitUI()
	libkb.G.UI.Configure()

	if localUser == "" {
		libkb.G.ConfigureSocketInfo()
		k, err := NewKBPKIClient(libkb.G)
		if err != nil {
			return nil, fmt.Errorf("Could not get KBPKI: %v", err)
		}
		config.SetKBPKI(k)

		c, err := NewCryptoClient(config.Codec(), libkb.G)
		if err != nil {
			return nil, fmt.Errorf("Could not get Crypto: %v", err)
		}
		config.SetCrypto(c)

		return config, nil
	}

	// localUser != ""

	users := []string{"strib", "max", "chris", "fred"}
	userIndex := -1
	for i := range users {
		if localUser == users[i] {
			userIndex = i
			break
		}
	}
	if userIndex < 0 {
		return nil, fmt.Errorf("user %s not in list %v", localUser, users)
	}

	localUsers := MakeLocalUsers(users)

	// TODO: Auto-generate these, too?
	localUsers[0].Asserts = []string{"github:strib"}
	localUsers[1].Asserts = []string{"twitter:maxtaco"}
	localUsers[2].Asserts = []string{"twitter:malgorithms"}
	localUsers[3].Asserts = []string{"twitter:fakalin"}

	var localUID keybase1.UID
	if userIndex >= 0 {
		localUID = localUsers[userIndex].UID
	}

	k := NewKBPKILocal(localUID, localUsers)
	config.SetKBPKI(k)

	signingKey := MakeLocalUserSigningKeyOrBust(localUser)
	cryptPrivateKey := MakeLocalUserCryptPrivateKeyOrBust(localUser)
	config.SetCrypto(NewCryptoLocal(config.Codec(), signingKey, cryptPrivateKey))

	return config, nil
}

// Shutdown does any necessary shutdown tasks for libkbfs. Shutdown
// should be called at the end of main.
func Shutdown(memProfilePath string) error {
	pprof.StopCPUProfile()

	if memProfilePath != "" {
		// Let the GC/OS clean up the file handle.
		f, err := os.Create(memProfilePath)
		if err != nil {
			return err
		}

		pprof.WriteHeapProfile(f)
	}

	return nil
}
