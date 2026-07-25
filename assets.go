package sharkie

import "embed"

//go:embed data/* winres/*
var Assets embed.FS
