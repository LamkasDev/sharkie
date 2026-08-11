module github.com/LamkasDev/sharkie

go 1.26.1

require (
	github.com/CovenantSQL/HashStablePack v2.0.0+incompatible
	github.com/LamkasDev/cimgui-go-vulkan v0.0.0-20260707145319-90e11755ff1b
	github.com/bpfsnoop/gapstone v0.0.0-20250326154852-7e3bee2c2f09
	github.com/cespare/xxhash v1.1.0
	github.com/ebitengine/oto/v3 v3.4.0
	github.com/elokore/glfw/v3.4/glfw v0.0.0-20251221231958-c1dc85df2170
	github.com/foize/go.fifo v0.0.0-20130327144150-3a04cfeec121
	github.com/goki/vulkan v1.0.8
	github.com/gookit/color v1.6.0
	github.com/langhuihui/gomem v0.0.0-20251013004544-ee5a2a75c165
	github.com/muesli/go-app-paths v0.2.2
	github.com/x448/float16 v0.8.4
	github.com/xlab/closer v1.1.0
	go.uber.org/atomic v1.11.0
	go101.org/nstd v0.2.3
	golang.org/x/sys v0.44.0
)

require (
	github.com/ebitengine/purego v0.9.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
)

replace github.com/langhuihui/gomem => ./temp/gomem

replace github.com/LamkasDev/cimgui-go-vulkan => ./temp/cimgui-go-vulkan
