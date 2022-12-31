package main

import (
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/utilitycmd/consts"
)

func main() {
	consts.InSpawnMode = os.Getenv(consts.UtilitySpawnEnvVar) == consts.UtilitySpawnEnvVarValue
	if consts.EnableDevMode {
		consts.InDevMode = os.Getenv("HPCGAME_DEV_MODE") == "true"
	}
	if !consts.InSpawnMode {
		if !consts.InDevMode {
			if os.Geteuid() != 0 {
				log.Fatalln(
					"Please ensure this executable is owned by root:root and have the setuid bit set.\r\n",
					"If under development, please export HPCGAME_DEV_MODE=true.\r\n",
					"Otherwise, please: chown root:root "+os.Args[0]+" && chmod u+s "+os.Args[0],
				)
			}
		}
	}
}
