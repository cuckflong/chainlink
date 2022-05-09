package keeper

import (
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/smartcontractkit/chainlink/core/services/job"
)

func ValidatedKeeperSpec(tomlString string) (job.Job, error) {
	// Create a new job with a randomly generated uuid, which can be replaced with the one from tomlString.
	var j = job.Job{
		ExternalJobID: uuid.NewV4(),
	}

	tree, err := toml.Load(tomlString)
	if err != nil {
		return j, err
	}

	if err := tree.Unmarshal(&j); err != nil {
		return j, err
	}

	var spec job.KeeperSpec
	if err := tree.Unmarshal(&spec); err != nil {
		return j, err
	}
	j.KeeperSpec = &spec

	if j.Type != job.Keeper {
		return j, errors.Errorf("unsupported type %s", j.Type)
	}

	return j, nil
}
