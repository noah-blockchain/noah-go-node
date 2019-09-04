package proxy

import (
	lite2 "github.com/noah-blockchain/noah-go-node/lite"
	"github.com/noah-blockchain/noah-go-node/lite/client"
	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func NewVerifier(chainID, rootDir string, client client.SignStatusClient, logger log.Logger, cacheSize int) (*lite2.DynamicVerifier, error) {

	logger = logger.With("module", "lite/proxy")
	logger.Info("lite/proxy/NewVerifier()...", "chainID", chainID, "rootDir", rootDir, "client", client)

	memProvider := lite2.NewDBProvider("trusted.mem", dbm.NewMemDB()).SetLimit(cacheSize)
	lvlProvider := lite2.NewDBProvider("trusted.lvl", dbm.NewDB("trust-base", dbm.GoLevelDBBackend, rootDir))
	trust := lite2.NewMultiProvider(
		memProvider,
		lvlProvider,
	)
	source := client.NewProvider(chainID, client)
	cert := lite2.NewDynamicVerifier(chainID, trust, source)
	cert.SetLogger(logger) // Sets logger recursively.

	// TODO: Make this more secure, e.g. make it interactive in the console?
	_, err := trust.LatestFullCommit(chainID, 1, 1<<63-1)
	if err != nil {
		logger.Info("lite/proxy/NewVerifier found no trusted full commit, initializing from source from height 1...")
		fc, err := source.LatestFullCommit(chainID, 1, 1)
		if err != nil {
			return nil, errors.Wrap(err, "fetching source full commit @ height 1")
		}
		err = trust.SaveFullCommit(fc)
		if err != nil {
			return nil, errors.Wrap(err, "saving full commit to trusted")
		}
	}

	return cert, nil
}
