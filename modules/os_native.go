package modules

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/robertkrimen/otto"
)

// setupOSNative installs an __os helper object on the VM with Go-backed
// functions that the os.js wrapper calls. This keeps OS interaction in
// Go while exposing a Node.js-compatible API through JavaScript.
func setupOSNative(vm *otto.Otto) {
	osObj, _ := vm.Object(`({})`)

	// --- Simple constants ---
	osObj.Set("eol", "\n")
	osObj.Set("devNull", "/dev/null")

	// --- Portable functions ---

	osObj.Set("hostname", func(call otto.FunctionCall) otto.Value {
		name, _ := os.Hostname()
		v, _ := otto.ToValue(name)
		return v
	})

	osObj.Set("homedir", func(call otto.FunctionCall) otto.Value {
		dir, _ := os.UserHomeDir()
		v, _ := otto.ToValue(dir)
		return v
	})

	osObj.Set("tmpdir", func(call otto.FunctionCall) otto.Value {
		v, _ := otto.ToValue(os.TempDir())
		return v
	})

	osObj.Set("platform", func(call otto.FunctionCall) otto.Value {
		v, _ := otto.ToValue(runtime.GOOS)
		return v
	})

	osObj.Set("arch", func(call otto.FunctionCall) otto.Value {
		a := runtime.GOARCH
		switch a {
		case "amd64":
			a = "x64"
		case "386":
			a = "ia32"
		}
		v, _ := otto.ToValue(a)
		return v
	})

	osObj.Set("type", func(call otto.FunctionCall) otto.Value {
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
		v, _ := otto.ToValue(t)
		return v
	})

	osObj.Set("endianness", func(call otto.FunctionCall) otto.Value {
		e := "LE"
		switch runtime.GOARCH {
		case "ppc64", "mips", "mips64", "s390x":
			e = "BE"
		}
		v, _ := otto.ToValue(e)
		return v
	})

	osObj.Set("availableParallelism", func(call otto.FunctionCall) otto.Value {
		v, _ := otto.ToValue(runtime.NumCPU())
		return v
	})

	// --- Linux syscall-backed functions ---

	var uname syscall.Utsname
	syscall.Uname(&uname)

	osObj.Set("release", func(call otto.FunctionCall) otto.Value {
		v, _ := otto.ToValue(byteFieldToString(uname.Release))
		return v
	})

	osObj.Set("version", func(call otto.FunctionCall) otto.Value {
		v, _ := otto.ToValue(byteFieldToString(uname.Version))
		return v
	})

	osObj.Set("machine", func(call otto.FunctionCall) otto.Value {
		v, _ := otto.ToValue(byteFieldToString(uname.Machine))
		return v
	})

	osObj.Set("uptime", func(call otto.FunctionCall) otto.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		v, _ := otto.ToValue(info.Uptime)
		return v
	})

	osObj.Set("freemem", func(call otto.FunctionCall) otto.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		v, _ := otto.ToValue(info.Freeram * uint64(info.Unit))
		return v
	})

	osObj.Set("totalmem", func(call otto.FunctionCall) otto.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		v, _ := otto.ToValue(info.Totalram * uint64(info.Unit))
		return v
	})

	osObj.Set("loadavg", func(call otto.FunctionCall) otto.Value {
		var info syscall.Sysinfo_t
		syscall.Sysinfo(&info)
		arr, _ := vm.Object(`([])`)
		arr.Call("push", float64(info.Loads[0])/65536.0)
		arr.Call("push", float64(info.Loads[1])/65536.0)
		arr.Call("push", float64(info.Loads[2])/65536.0)
		return arr.Value()
	})

	// --- CPUs (reads /proc/cpuinfo on Linux) ---

	osObj.Set("cpus", func(call otto.FunctionCall) otto.Value {
		arr, _ := vm.Object(`([])`)
		cpus := parseCPUInfo()
		for _, c := range cpus {
			cpu, _ := vm.Object(`({})`)
			cpu.Set("model", c.model)
			cpu.Set("speed", c.speed)
			times, _ := vm.Object(`({})`)
			times.Set("user", 0)
			times.Set("nice", 0)
			times.Set("sys", 0)
			times.Set("idle", 0)
			times.Set("irq", 0)
			cpu.Set("times", times.Value())
			arr.Call("push", cpu.Value())
		}
		return arr.Value()
	})

	// --- Network interfaces ---

	osObj.Set("networkInterfaces", func(call otto.FunctionCall) otto.Value {
		result, _ := vm.Object(`({})`)
		ifaces, err := net.Interfaces()
		if err != nil {
			return result.Value()
		}
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil || len(addrs) == 0 {
				continue
			}
			arr, _ := vm.Object(`([])`)
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				entry, _ := vm.Object(`({})`)
				entry.Set("address", ipnet.IP.String())
				entry.Set("netmask", net.IP(ipnet.Mask).String())
				family := "IPv4"
				if ipnet.IP.To4() == nil {
					family = "IPv6"
				}
				entry.Set("family", family)
				mac := iface.HardwareAddr.String()
				if mac == "" {
					mac = "00:00:00:00:00:00"
				}
				entry.Set("mac", mac)
				entry.Set("internal", iface.Flags&net.FlagLoopback != 0)
				ones, _ := ipnet.Mask.Size()
				entry.Set("cidr", fmt.Sprintf("%s/%d", ipnet.IP.String(), ones))
				arr.Call("push", entry.Value())
			}
			result.Set(iface.Name, arr.Value())
		}
		return result.Value()
	})

	// --- User info ---

	osObj.Set("userInfo", func(call otto.FunctionCall) otto.Value {
		info, _ := vm.Object(`({})`)
		u, err := user.Current()
		if err != nil {
			return info.Value()
		}
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		info.Set("uid", uid)
		info.Set("gid", gid)
		info.Set("username", u.Username)
		info.Set("homedir", u.HomeDir)
		shell := os.Getenv("SHELL")
		if shell == "" {
			info.Set("shell", otto.NullValue())
		} else {
			info.Set("shell", shell)
		}
		return info.Value()
	})

	vm.Set("__os", osObj)
}

// byteFieldToString converts a [65]int8 utsname field to a Go string.
func byteFieldToString(arr [65]int8) string {
	n := 0
	for n < len(arr) && arr[n] != 0 {
		n++
	}
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(arr[i])
	}
	return string(buf)
}

type cpuInfo struct {
	model string
	speed int
}

// parseCPUInfo reads /proc/cpuinfo and returns per-core info.
func parseCPUInfo() []cpuInfo {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		n := runtime.NumCPU()
		cpus := make([]cpuInfo, n)
		return cpus
	}

	var cpus []cpuInfo
	var cur cpuInfo
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cur.model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				f, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				cur.speed = int(f)
			}
		} else if line == "" && cur.model != "" {
			cpus = append(cpus, cur)
			cur = cpuInfo{}
		}
	}
	if cur.model != "" {
		cpus = append(cpus, cur)
	}

	// Fallback if parsing yielded nothing.
	if len(cpus) == 0 {
		cpus = make([]cpuInfo, runtime.NumCPU())
	}
	return cpus
}
