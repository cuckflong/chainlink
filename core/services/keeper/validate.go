package keeper

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink/core/services/job"
	"github.com/smartcontractkit/chainlink/core/services/pipeline"
	"strings"
)

const ObservationSource = `
    encode_check_upkeep_tx   [type=ethabiencode
                              abi="checkUpkeep(uint256 id, address from)"
                              data="{\"id\":$(jobSpec.upkeepID),\"from\":$(jobSpec.fromAddress)}"]
    check_upkeep_tx          [type=ethcall
                              failEarly=true
                              extractRevertReason=true
                              evmChainID="$(jobSpec.evmChainID)"
                              contract="$(jobSpec.contractAddress)"
                              gas="$(jobSpec.checkUpkeepGasLimit)"
                              gasPrice="$(jobSpec.gasPrice)"
                              gasTipCap="$(jobSpec.gasTipCap)"
                              gasFeeCap="$(jobSpec.gasFeeCap)"
                              data="$(encode_check_upkeep_tx)"]
    decode_check_upkeep_tx   [type=ethabidecode
                              abi="bytes memory performData, uint256 maxLinkPayment, uint256 gasLimit, uint256 adjustedGasWei, uint256 linkEth"]
    encode_perform_upkeep_tx [type=ethabiencode
                              abi="performUpkeep(uint256 id, bytes calldata performData)"
                              data="{\"id\": $(jobSpec.upkeepID),\"performData\":$(decode_check_upkeep_tx.performData)}"]
    perform_upkeep_tx        [type=ethtx
                              minConfirmations=0
                              to="$(jobSpec.contractAddress)"
                              from="[$(jobSpec.fromAddress)]"
                              evmChainID="$(jobSpec.evmChainID)"
                              data="$(encode_perform_upkeep_tx)"
                              gasLimit="$(jobSpec.performUpkeepGasLimit)"
                              txMeta="{\"jobID\":$(jobSpec.jobID),\"upkeepID\":$(jobSpec.prettyID)}"]
    encode_check_upkeep_tx -> check_upkeep_tx -> decode_check_upkeep_tx -> encode_perform_upkeep_tx -> perform_upkeep_tx
`

var parsedPipeline pipeline.Pipeline

// We parse the ObservationSource only once here, because it is constant for all the Keeper jobs.
func init() {
	parsed, err := pipeline.Parse(ObservationSource)

	if err != nil {
		panic(fmt.Sprintf("Failed to parse default Keeper observation source: %v", err))
	}

	parsedPipeline = *parsed
}

func ValidatedKeeperSpec(tomlString string) (job.Job, error) {
	// Create a new job with a randomly generated uuid, which can be replaced with the one from tomlString.
	var j = job.Job{
		ExternalJobID: uuid.NewV4(),
	}

	if strings.Contains(tomlString, "observationSource") {
		return j, errors.New("observationSource not required in the toml string")
	}

	j.Pipeline = parsedPipeline

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
