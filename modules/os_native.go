package modules

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"syscall"

	"github.com/dop251/goja"
)

// setupOSNative installs an __os helper object on the VM with Go-backed
// functions that the os.js wrapper calls.
func setupOSNative(vm *goja.Runtime) {
	osObj := vm.NewObject()

	// --- Simple constants ---
	osObj.Set("eol", "\n")
	osObj.Set("devNull", "/dev/null")

	// --- Portable functions ---

	osObj.Set("hostname", func(call goja.FunctionCall) goja.Value {
		name, _ := os.Hostname()
		return vm.ToValue(name)
	})

	osObj.Set("homedir", func(call goja.FunctionCall) goja.Value {
		dir, _ := os.UserHomeDir()
		return vm.ToValue(dir)
	})

	osObj.Set("tmpdir", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.TempDir())
	})

	osObj.Set("platform", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(runtime.GOOS)
	})

	osObj.Set("arch", func(call goja.FunctionCall) goja.Value {
		a := runtime.GOARCH
		switch a {
		case "amd64":
			a = "x64"
		case "386":
			a = "ia32"
		}
		return vm.ToValue(a)
	})

	osObj.Set("type", func(call goja.FunctionCall) goja.Value {
		t := runtime.GOOS
		switch t {
		case "linux":
			t = "Linux"
		case "darwin":
			t = "Darwin"
		case "windows":
			t = "Windows_NT"
		case "freebsd":
			t = "FreeBSD"
		}
		return vm.ToValue(t)
	})

	osObj.Set("endianness", func(call goja.FunctionCall) goja.Value {
		e := "LE"
		switch runtime.GOARCH {
		case "ppc64", "mips", "mips64", "s390x":
			e = "BE"
		}
		return vm.ToValue(e)
	})

	osObj.Set("availableParallelism", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(runtime.NumCPU())
	})

	// --- Linux syscall-backed functions ---

	var uname syscall.Utsname
	syscall.Uname(&uname)

	osObj.Set("release", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(byteFieldToString(uname.Release))
	})

	osObj.Set("version", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(byteFieldToString(uname.Version))
	})

	osObj.Set("machine", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(byteFieldToString(uname.Machine))
	})

	osObj.Set("uptime", func(call goja.FunctionCall) goja.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		return vm.ToValue(info.Uptime)
	})

	osObj.Set("freemem", func(call goja.FunctionCall) goja.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		return vm.ToValue(info.Freeram * uint64(info.Unit))
	})

	osObj.Set("totalmem", func(call goja.FunctionCall) goja.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		return vm.ToValue(info.Totalram * uint64(info.Unit))
	})

	osObj.Set("loadavg", func(call goja.FunctionCall) goja.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		return vm.ToValue([]float64{
			float64(info.Loads[0]) / 65536.0,
			float64(info.Loads[1]) / 65536.0,
			float64(info.Loads[2]) / 65536.0,
		})
	})

	// --- CPUs ---

	osObj.Set("cpus", func(call goja.FunctionCall) goja.Value {
		cpus := parseCPUInfo()
		result := make([]interface{}, len(cpus))
		for i, c := range cpus {
			times := map[string]interface{}{
				"user": 0, "nice": 0, "sys": 0, "idle": 0, "irq": 0,
			}
			result[i] = map[string]interface{}{
				"model": c.model, "speed": c.speed, "times": times,
			}
		}
		return vm.ToValue(result)
	})

	// --- Network interfaces ---

	osObj.Set("networkInterfaces", func(call goja.FunctionCall) goja.Value {
		result := vm.NewObject()
		ifaces, err := net.Interfaces()
		if err != nil {
			return result
		}
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil || len(addrs) == 0 {
				continue
			}
			var entries []interface{}
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				family := "IPv4"
				if ipnet.IP.To4() == nil {
					family = "IPv6"
				}
				mac := iface.HardwareAddr.String()
				if mac == "" {
					mac = "00:00:00:00:00:00"
				}
				ones, _ := ipnet.Mask.Size()
				entry := map[string]interface{}{
					"address":  ipnet.IP.String(),
					"netmask":  net.IP(ipnet.Mask).String(),
					"family":   family,
					"mac":      mac,
					"internal": iface.Flags&net.FlagLoopback != 0,
					"cidr":     fmt.Sprintf("%s/%d", ipnet.IP.String(), ones),
				}
				entries = append(entries, entry)
			}
			result.Set(iface.Name, entries)
		}
		return result
	})

	// --- User info ---

	osObj.Set("userInfo", func(call goja.FunctionCall) goja.Value {
		info := vm.NewObject()
		u, err := user.Current()
		if err != nil {
			return info
		}
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		info.Set("uid", uid)
		info.Set("gid", gid)
		info.Set("username", u.Username)
		info.Set("homedir", u.HomeDir)
		shell := os.Getenv("SHELL")
		if shell == "" {
			info.Set("shell", goja.Null())
		} else {
			info.Set("shell", shell)
		}
		return info
	})

	vm.Set("__os", osObj)
}
