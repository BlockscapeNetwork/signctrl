package init

import (
	"fmt"
	"io/ioutil"
	"os"
)

const configTemplate = `[init]

# The minimum log level for pairmint logs.
# Must be either DEBUG, INFO, WARN or ERR.
log_level = ""

# The number of pairminted validator nodes run in parallel.
# This number cannot be changed during operation anymore.
# If you want to change the set_size you need to stop all
# nodes, modify the configuration on each node and start
# them.
# Must be 2 or higher.
set_size = <int>

# The number of missed blocks in a row that trigger a rank
# update in the set.
# Must be 1 or higher.
threshold = <int>

# The nodes rank on startup. It is used to determine which node
# in the set is currently signing and which nodes line up as
# backups. Rank 1 is always the signer, while ranke 2 and above
# move up one rank each time the threshold of missed blocks in
# a row is reached and thus a new signer is determined.
# Must be 1 or higher.
rank = <int>

[connection]

# The TCP socket address of the validator node for Pairmint to
# connect to.
validator_addr = "tcp://127.0.0.1:26658"

# The TCP socket address for Pairmint to listen on for connections
# from an external PrivValidator process.
priv_validator_laddr = "tcp://"
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
