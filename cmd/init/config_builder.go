package init

import (
	"fmt"
	"io/ioutil"
	"os"
)

const configTemplate = `[init]

# Minimum log level for Pairmint logs.
# Must be either DEBUG, INFO, WARN or ERR.
log_level = ""

# Number of pairminted validator nodes running in parallel.
# This number cannot be changed during operation anymore.
# If you want to change this value you need to stop all
# nodes, modify the configuration on each node and start
# them up again.
# Must be 2 or higher.
set_size = 0

# Number of missed blocks in a row that triggers a rank
# update in the set.
# Must be 1 or higher.
threshold = 0

# Node's rank on startup. It is used to determine which node
# in the set is currently signing (rank 1) and which nodes
# line up as backups (ranks 2-n).
# Must be 1 or higher.
rank = 0

# TCP socket address the validator listens on for an external
# PrivValidator process. Pairmint dials this address to
# establish a connection to the validator.
# Must be in host:port format.
validator_laddr = ""

# TCP socket address the validator's RPC server listens on.
# Must be in host:port format.
validator_laddr_rpc = ""

[file_pv]

# The chain ID the signer should sign for.
chain_id = ""

# The path to the priv_validator_key.json file.
key_file_path = ""

# The path to the priv_validator_state.json file.
state_file_path = ""
`

// BuildConfigTemplate creates a pairmint.toml with a configuration template.
func BuildConfigTemplate(configDir string) error {
	if _, err := os.Stat(configDir + "/pairmint.toml"); !os.IsNotExist(err) {
		fmt.Printf("Found existing pairmint.toml at %v\n", configDir)
		return nil
	}
	if err := ioutil.WriteFile(configDir+"/pairmint.toml", []byte(configTemplate), 0644); err != nil {
		return err
	}
	fmt.Printf("Created new pairmint.toml template at %v\n", configDir)

	return nil
}
