package main

import (
	"flag"

	"github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// Constants
const htmlAbout = `Welcome on <b>Astilectron</b> demo!<br>
This is using the bootstrap and the bundler.`

// Vars
var (
	AppName string
	BuiltAt string
	debug   = flag.Bool("d", false, "enables the debug mode")
	w       *astilectron.Window
)

func main() {
	// Init
	flag.Parse()
	astilog.FlagInit()

	// Run bootstrap
	astilog.Debugf("Running app built at %s", BuiltAt)
	if err := bootstrap.Run(bootstrap.Options{
		Asset:    Asset,
		AssetDir: AssetDir,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/icon.icns",
			AppIconDefaultPath: "resources/icon.png",
		},
		Debug: *debug,
		MenuOptions: []*astilectron.MenuItemOptions{{
			Label: astilectron.PtrStr("Edit"),
			SubMenu: []*astilectron.MenuItemOptions{
				{
					Label:       astilectron.PtrStr("Cut"),
					Accelerator: astilectron.NewAccelerator("CmdOrCtrl+X"),
					Role:        astilectron.PtrStr("cut"),
				},
				{
					Label:       astilectron.PtrStr("Copy"),
					Accelerator: astilectron.NewAccelerator("CmdOrCtrl+C"),
					Role:        astilectron.PtrStr("copy"),
				},
				{
					Label:       astilectron.PtrStr("Paste"),
					Accelerator: astilectron.NewAccelerator("CmdOrCtrl+V"),
					Role:        astilectron.PtrStr("paste"),
				},
				{
					Label:       astilectron.PtrStr("Select All"),
					Accelerator: astilectron.NewAccelerator("CmdOrCtrl+A"),
					Role:        astilectron.PtrStr("selectall"),
				},
			},
		}},
		RestoreAssets: RestoreAssets,
		Windows: []*bootstrap.Window{{
			Homepage:       "index.html",
			MessageHandler: handleMessages,
			Options: &astilectron.WindowOptions{
				BackgroundColor: astilectron.PtrStr("#333"),
				Center:          astilectron.PtrBool(true),
				MinHeight:       astilectron.PtrInt(680),
				MinWidth:        astilectron.PtrInt(415),
				Height:          astilectron.PtrInt(680),
				Width:           astilectron.PtrInt(415),
			},
		}},
	}); err != nil {
		astilog.Fatal(errors.Wrap(err, "running bootstrap failed"))
	}
}
