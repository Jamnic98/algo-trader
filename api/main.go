package main

import (
	"api/setup"
)

func main() {
  cfg := setup.GetConfig()
  setup.InitProject(cfg)
}
