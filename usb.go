package main

// import (
// 	"flag"
// 	"fmt"
// 	"log"

// 	"github.com/google/gousb"
// 	"github.com/google/gousb/usbid"
// 	"google.golang.org/api/gmail/v1"
// )

// func detectusb(srv *gmail.Service) {
// 	flag.Parse()

// 	// Only one context should be needed for an application.  It should always be closed.
// 	ctx := gousb.NewContext()
// 	defer ctx.Close()

// 	// Debugging can be turned on; this shows some of the inner workings of the libusb package.
// 	// ctx.Debug(*debug)

// 	// OpenDevices is used to find the devices to open.
// 	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
// 		// The usbid package can be used to print out human readable information.
// 		fmt.Printf("%03d.%03d %s:%s %s\n", desc.Bus, desc.Address, desc.Vendor, desc.Product, usbid.Describe(desc))
// 		fmt.Printf("  Protocol: %s\n", usbid.Classify(desc))

// 		// The configurations can be examined from the DeviceDesc, though they can only
// 		// be set once the device is opened.  All configuration references must be closed,
// 		// to free up the memory in libusb.
// 		for _, cfg := range desc.Configs {
// 			// This loop just uses more of the built-in and usbid pretty printing to list
// 			// the USB devices.
// 			fmt.Printf("  %s:\n", cfg)
// 			for _, intf := range cfg.Interfaces {
// 				fmt.Printf("    --------------\n")
// 				for _, ifSetting := range intf.AltSettings {
// 					fmt.Printf("    %s\n", ifSetting)
// 					fmt.Printf("      %s\n", usbid.Classify(ifSetting))
// 					for _, end := range ifSetting.Endpoints {
// 						fmt.Printf("      %s\n", end)
// 					}
// 				}
// 			}
// 			fmt.Printf("    --------------\n")
// 		}

// 		// After inspecting the descriptor, return true or false depending on whether
// 		// the device is "interesting" or not.  Any descriptor for which true is returned
// 		// opens a Device which is retuned in a slice (and must be subsequently closed).
// 		return false
// 	})

// 	// All Devices returned from OpenDevices must be closed.
// 	defer func() {
// 		for _, d := range devs {
// 			d.Close()
// 		}
// 	}()

// 	// OpenDevices can occasionally fail, so be sure to check its return value.
// 	if err != nil {
// 		log.Fatalf("list: %s", err)
// 	}
// }
