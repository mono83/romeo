package services

import (
	"github.com/mono83/romeo"
)

type nameHolder string

func (n nameHolder) GetName() string           { return string(n) }
func (nameHolder) GetRunLevel() romeo.RunLevel { return romeo.RunLevelMain }
