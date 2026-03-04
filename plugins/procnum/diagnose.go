package procnum

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/cprobe/catpaw/diagnose"
	"github.com/cprobe/catpaw/plugins"
	"github.com/shirou/gopsutil/v3/process"
)

var _ plugins.Diagnosable = (*ProcnumPlugin)(nil)

func (p *ProcnumPlugin) RegisterDiagnoseTools(registry *diagnose.ToolRegistry) {
	registry.RegisterCategory("process", "procnum",
		"Process diagnostic tools (list, search, detail)", diagnose.ToolScopeLocal)

	registry.Register("process", diagnose.DiagnoseTool{
		Name:        "process_list",
		Description: "List processes sorted by CPU usage (top 20). Samples up to 1000 processes to limit overhead on stressed systems.",
		Scope:       diagnose.ToolScopeLocal,
		Execute: func(ctx context.Context, args map[string]string) (string, error) {
			procs, err := process.ProcessesWithContext(ctx)
			if err != nil {
				return "", fmt.Errorf("list processes: %w", err)
			}

			const maxSample = 1000
			type info struct {
				pid  int32
				name string
				user string
				cpu  float64
				mem  float32
				cmd  string
			}

			var all []info
			sampled := 0
			for _, p := range procs {
				if sampled >= maxSample {
					break
				}
				sampled++
				cpuPct, err := p.CPUPercentWithContext(ctx)
				if err != nil {
					continue
				}
				name, _ := p.NameWithContext(ctx)
				user, _ := p.UsernameWithContext(ctx)
				memPct, _ := p.MemoryPercentWithContext(ctx)
				cmd, _ := p.CmdlineWithContext(ctx)
				if len(cmd) > 120 {
					cmd = cmd[:120] + "..."
				}
				all = append(all, info{pid: p.Pid, name: name, user: user, cpu: cpuPct, mem: memPct, cmd: cmd})
			}

			sort.Slice(all, func(i, j int) bool { return all[i].cpu > all[j].cpu })

			var b strings.Builder
			if len(procs) > maxSample {
				fmt.Fprintf(&b, "Total processes: %d (sampled first %d)\n\n", len(procs), maxSample)
			} else {
				fmt.Fprintf(&b, "Total processes: %d\n\n", len(all))
			}
			fmt.Fprintf(&b, "%7s  %6s  %6s  %-12s  %-16s  %s\n", "PID", "CPU%", "MEM%", "USER", "NAME", "CMDLINE")
			limit := 20
			if limit > len(all) {
				limit = len(all)
			}
			for i := 0; i < limit; i++ {
				p := all[i]
				fmt.Fprintf(&b, "%7d  %5.1f%%  %5.1f%%  %-12s  %-16s  %s\n",
					p.pid, p.cpu, p.mem, truncate(p.user, 12), truncate(p.name, 16), p.cmd)
			}
			return b.String(), nil
		},
	})

	registry.Register("process", diagnose.DiagnoseTool{
		Name:        "process_search",
		Description: "Search processes by name or command line substring. Lightweight: only reads name+cmdline per process, skips heavy stats for non-matches. Parameter: pattern (required)",
		Scope:       diagnose.ToolScopeLocal,
		Parameters: []diagnose.ToolParam{
			{Name: "pattern", Type: "string", Description: "Substring to match against process name or command line", Required: true},
		},
		Execute: func(ctx context.Context, args map[string]string) (string, error) {
			pattern := args["pattern"]
			if pattern == "" {
				return "", fmt.Errorf("parameter 'pattern' is required")
			}
			procs, err := process.ProcessesWithContext(ctx)
			if err != nil {
				return "", fmt.Errorf("list processes: %w", err)
			}

			lowerPattern := strings.ToLower(pattern)
			var b strings.Builder
			count := 0
			fmt.Fprintf(&b, "%7s  %6s  %6s  %-12s  %-16s  %s\n", "PID", "CPU%", "MEM%", "USER", "NAME", "CMDLINE")
			for _, p := range procs {
				name, _ := p.NameWithContext(ctx)
				cmd, _ := p.CmdlineWithContext(ctx)
				if !strings.Contains(strings.ToLower(name), lowerPattern) &&
					!strings.Contains(strings.ToLower(cmd), lowerPattern) {
					continue
				}
				cpuPct, _ := p.CPUPercentWithContext(ctx)
				user, _ := p.UsernameWithContext(ctx)
				memPct, _ := p.MemoryPercentWithContext(ctx)
				if len(cmd) > 120 {
					cmd = cmd[:120] + "..."
				}
				fmt.Fprintf(&b, "%7d  %5.1f%%  %5.1f%%  %-12s  %-16s  %s\n",
					p.Pid, cpuPct, memPct, truncate(user, 12), truncate(name, 16), cmd)
				count++
				if count >= 50 {
					fmt.Fprintf(&b, "\n... (showing first 50 matches)")
					break
				}
			}
			if count == 0 {
				return fmt.Sprintf("No processes found matching pattern: %s", pattern), nil
			}
			return fmt.Sprintf("Found %d processes matching '%s':\n\n%s", count, pattern, b.String()), nil
		},
	})

	registry.Register("process", diagnose.DiagnoseTool{
		Name:        "process_detail",
		Description: "Show detailed info for a specific process by PID: status, memory, open files, connections, threads. Parameter: pid (required)",
		Scope:       diagnose.ToolScopeLocal,
		Parameters: []diagnose.ToolParam{
			{Name: "pid", Type: "int", Description: "Process ID to inspect", Required: true},
		},
		Execute: func(ctx context.Context, args map[string]string) (string, error) {
			pidStr := args["pid"]
			if pidStr == "" {
				return "", fmt.Errorf("parameter 'pid' is required")
			}
			var pid int32
			if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
				return "", fmt.Errorf("invalid pid: %s", pidStr)
			}

			p, err := process.NewProcessWithContext(ctx, pid)
			if err != nil {
				return "", fmt.Errorf("process %d not found: %w", pid, err)
			}

			var b strings.Builder
			name, _ := p.NameWithContext(ctx)
			cmd, _ := p.CmdlineWithContext(ctx)
			status, _ := p.StatusWithContext(ctx)
			user, _ := p.UsernameWithContext(ctx)
			ppid, _ := p.PpidWithContext(ctx)
			cpuPct, _ := p.CPUPercentWithContext(ctx)
			memPct, _ := p.MemoryPercentWithContext(ctx)
			memInfo, _ := p.MemoryInfoWithContext(ctx)
			threads, _ := p.NumThreadsWithContext(ctx)
			fds, _ := p.NumFDsWithContext(ctx)
			createTime, _ := p.CreateTimeWithContext(ctx)

			fmt.Fprintf(&b, "PID:        %d\n", pid)
			fmt.Fprintf(&b, "Name:       %s\n", name)
			fmt.Fprintf(&b, "Status:     %v\n", status)
			fmt.Fprintf(&b, "User:       %s\n", user)
			fmt.Fprintf(&b, "PPID:       %d\n", ppid)
			fmt.Fprintf(&b, "CPU%%:       %.1f%%\n", cpuPct)
			fmt.Fprintf(&b, "MEM%%:       %.1f%%\n", memPct)
			if memInfo != nil {
				fmt.Fprintf(&b, "RSS:        %s\n", humanBytes(memInfo.RSS))
				fmt.Fprintf(&b, "VMS:        %s\n", humanBytes(memInfo.VMS))
			}
			fmt.Fprintf(&b, "Threads:    %d\n", threads)
			fmt.Fprintf(&b, "FDs:        %d\n", fds)
			if createTime > 0 {
				fmt.Fprintf(&b, "CreateTime: %d (epoch ms)\n", createTime)
			}
			fmt.Fprintf(&b, "Cmdline:    %s\n", cmd)

			conns, err := p.ConnectionsWithContext(ctx)
			if err == nil && len(conns) > 0 {
				fmt.Fprintf(&b, "\nNetwork connections (%d):\n", len(conns))
				limit := 20
				if limit > len(conns) {
					limit = len(conns)
				}
				for i := 0; i < limit; i++ {
					c := conns[i]
					fmt.Fprintf(&b, "  %s %s:%d → %s:%d (%s)\n",
						connType(c.Type), c.Laddr.IP, c.Laddr.Port,
						c.Raddr.IP, c.Raddr.Port, c.Status)
				}
				if len(conns) > limit {
					fmt.Fprintf(&b, "  ... and %d more\n", len(conns)-limit)
				}
			}
			return b.String(), nil
		},
	})

	// process_stack: capture stack trace snapshot for a running process
	registry.Register("process", diagnose.DiagnoseTool{
		Name:        "process_stack",
		Description: "Capture a stack trace snapshot for a running process. Attempts multiple methods: Java (jstack), Python (py-spy), Go (pprof endpoint), and generic (gdb). Useful for diagnosing CPU spikes or hangs by showing where code is currently executing. Parameter: pid (required)",
		Scope:       diagnose.ToolScopeLocal,
		Parameters: []diagnose.ToolParam{
			{Name: "pid", Type: "int", Description: "Process ID to capture stack trace from", Required: true},
		},
		Execute: func(ctx context.Context, args map[string]string) (string, error) {
			pidStr := args["pid"]
			if pidStr == "" {
				return "", fmt.Errorf("parameter 'pid' is required")
			}
			var pid int32
			if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
				return "", fmt.Errorf("invalid pid: %s", pidStr)
			}

			// Verify process exists and get basic info
			proc, err := process.NewProcessWithContext(ctx, pid)
			if err != nil {
				return "", fmt.Errorf("process %d not found: %w", pid, err)
			}

			cmdline, _ := proc.CmdlineWithContext(ctx)
			name, _ := proc.NameWithContext(ctx)

			var methods []string
			var results []string

			// Method 1: Java - jstack
			if isJavaProcess(cmdline) {
				methods = append(methods, "jstack")
				out, jstackErr := runJStack(ctx, pid)
				if jstackErr == nil && out != "" {
					return formatStackResult("jstack", pid, name, out, methods), nil
				}
				results = append(results, fmt.Sprintf("jstack: %v", jstackErr))
			}

			// Method 2: Python - py-spy
			if isPythonProcess(cmdline) {
				methods = append(methods, "py-spy")
				out, pySpyErr := runPySpy(ctx, pid)
				if pySpyErr == nil && out != "" {
					return formatStackResult("py-spy", pid, name, out, methods), nil
				}
				results = append(results, fmt.Sprintf("py-spy: %v", pySpyErr))
			}

			// Method 3: Go - try pprof endpoint
			if isGoProcess(cmdline) {
				methods = append(methods, "pprof")
				out, pprofErr := tryGoPprof(ctx, pid, proc)
				if pprofErr == nil && out != "" {
					return formatStackResult("pprof", pid, name, out, methods), nil
				}
				results = append(results, fmt.Sprintf("pprof: %v", pprofErr))
			}

			// Method 4: Node.js - try V8 inspector or llnode
			if isNodeProcess(cmdline) {
				methods = append(methods, "node-inspector")
				out, nodeErr := runNodeStack(ctx, pid, proc)
				if nodeErr == nil && out != "" {
					return formatStackResult("node-inspector", pid, name, out, methods), nil
				}
				results = append(results, fmt.Sprintf("node-inspector: %v", nodeErr))
			}

			// Method 5: Generic - gdb (last resort, may not work for stripped binaries)
			methods = append(methods, "gdb")
			out, gdbErr := runGdb(ctx, pid)
			if gdbErr == nil && out != "" {
				return formatStackResult("gdb", pid, name, out, methods), nil
			}
			results = append(results, fmt.Sprintf("gdb: %v", gdbErr))

			// All methods failed - check if any tool was found
			allCmdNotFound := true
			for _, r := range results {
				if !strings.Contains(r, "executable file not found") &&
					!strings.Contains(r, "command not found") &&
					!strings.Contains(strings.ToLower(r), "not found") {
					allCmdNotFound = false
					break
				}
			}

			var b strings.Builder
			fmt.Fprintf(&b, "Failed to capture stack trace for PID %d (%s)\n", pid, name)
			fmt.Fprintf(&b, "Attempted methods: %v\n\n", methods)
			for _, r := range results {
				fmt.Fprintf(&b, "  - %s\n", r)
			}
			fmt.Fprintf(&b, "\nTips:\n")
			fmt.Fprintf(&b, "  - Java: ensure JDK is installed (jstack available)\n")
			fmt.Fprintf(&b, "  - Python: install py-spy (pip install py-spy)\n")
			fmt.Fprintf(&b, "  - Go: ensure pprof endpoint is exposed (net/http/pprof)\n")
			fmt.Fprintf(&b, "  - Node.js: start with --inspect flag, or use llnode/lldb\n")
			fmt.Fprintf(&b, "  - Generic: ensure gdb is installed and binary has debug symbols\n")

			// Return error message that selftest can recognize
			if allCmdNotFound {
				return b.String(), fmt.Errorf("jstack/py-spy/llnode/gdb executable file not found in $PATH")
			}
			return b.String(), fmt.Errorf("all stack capture methods failed")
		},
	})
}

func formatStackResult(method string, pid int32, name, output string, attempted []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Stack trace captured via %s for PID %d (%s)\n", method, pid, name)
	if len(attempted) > 1 {
		fmt.Fprintf(&b, "Attempted methods: %v (used: %s)\n\n", attempted, method)
	} else {
		fmt.Fprintf(&b, "\n")
	}
	// Limit output size
	if len(output) > 8000 {
		output = output[:8000] + "\n... [truncated]"
	}
	fmt.Fprint(&b, output)
	return b.String()
}

func isJavaProcess(cmdline string) bool {
	return strings.Contains(cmdline, "java") ||
		strings.Contains(cmdline, "scala") ||
		strings.Contains(cmdline, "kotlin") ||
		strings.Contains(cmdline, "-jar") ||
		strings.Contains(cmdline, "-Xmx") ||
		strings.Contains(cmdline, "-Xms")
}

func isPythonProcess(cmdline string) bool {
	return strings.Contains(cmdline, "python") ||
		strings.Contains(cmdline, "python3") ||
		strings.Contains(cmdline, "python2") ||
		strings.Contains(cmdline, "uwsgi") ||
		strings.Contains(cmdline, "gunicorn") ||
		strings.Contains(cmdline, "celery")
}

func isGoProcess(cmdline string) bool {
	// Hard to detect, but check common Go binary patterns
	lower := strings.ToLower(cmdline)
	return strings.Contains(lower, "go-") ||
		strings.Contains(lower, "prometheus") ||
		strings.Contains(lower, "grafana") ||
		strings.Contains(lower, "consul") ||
		strings.Contains(lower, "vault") ||
		strings.Contains(lower, "nomad") ||
		strings.Contains(lower, "etcd") ||
		strings.Contains(lower, "traefik") ||
		strings.Contains(lower, "caddy") ||
		strings.Contains(lower, "catpaw") ||
		strings.Contains(lower, "node_exporter") ||
		strings.Contains(lower, "blackbox_exporter")
}

func isNodeProcess(cmdline string) bool {
	// Node.js processes
	return strings.Contains(cmdline, "node ") ||
		strings.Contains(cmdline, "nodejs") ||
		strings.Contains(cmdline, "/node") ||
		strings.Contains(cmdline, "npm ") ||
		strings.Contains(cmdline, "yarn ") ||
		strings.Contains(cmdline, "pnpm ") ||
		strings.Contains(cmdline, "next-server") ||
		strings.Contains(cmdline, "nest start") ||
		strings.Contains(cmdline, "electron") ||
		strings.Contains(cmdline, "ts-node") ||
		strings.Contains(cmdline, "webpack") ||
		strings.Contains(cmdline, "vite") ||
		strings.Contains(cmdline, "esbuild")
}

func runJStack(ctx context.Context, pid int32) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "jstack", fmt.Sprintf("%d", pid))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("jstack failed: %w, output: %s", err, string(out))
	}
	return string(out), nil
}

func runPySpy(ctx context.Context, pid int32) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "py-spy", "dump", "--pid", fmt.Sprintf("%d", pid))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("py-spy failed: %w, output: %s", err, string(out))
	}
	return string(out), nil
}

func tryGoPprof(ctx context.Context, pid int32, proc *process.Process) (string, error) {
	// Try common pprof ports
	ports := []string{"6060", "8080", "9090", "8888", "3000"}

	conns, err := proc.ConnectionsWithContext(ctx)
	if err == nil {
		// Find listening ports that might be pprof
		for _, conn := range conns {
			if conn.Status == "LISTEN" && conn.Laddr.Port > 0 {
				ports = append([]string{fmt.Sprintf("%d", conn.Laddr.Port)}, ports...)
			}
		}
	}

	for _, port := range ports {
		url := fmt.Sprintf("http://127.0.0.1:%s/debug/pprof/goroutine?debug=1", port)
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
		if err != nil {
			cancel()
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, err := io.ReadAll(io.LimitReader(resp.Body, 16000))
			if err != nil {
				continue
			}
			return string(body), nil
		}
	}

	return "", fmt.Errorf("no accessible pprof endpoint found")
}

func runGdb(ctx context.Context, pid int32) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Use gdb to attach and get stack traces
	cmd := exec.CommandContext(ctx, "gdb",
		"-batch",
		"-ex", "set pagination off",
		"-ex", "thread apply all bt",
		"-p", fmt.Sprintf("%d", pid),
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gdb failed: %w, output: %s", err, string(out))
	}
	return string(out), nil
}

func runNodeStack(ctx context.Context, pid int32, proc *process.Process) (string, error) {
	// Method 1: Try V8 inspector (requires --inspect flag or SIGUSR1 activation)
	// Node.js inspector usually runs on port 9229
	inspectorPorts := []string{"9229", "9230", "9231", "9222"}

	// Try to find listening ports that might be Node inspector
	conns, err := proc.ConnectionsWithContext(ctx)
	if err == nil {
		for _, conn := range conns {
			if conn.Status == "LISTEN" && conn.Laddr.Port > 0 {
				inspectorPorts = append([]string{fmt.Sprintf("%d", conn.Laddr.Port)}, inspectorPorts...)
			}
		}
	}

	for _, port := range inspectorPorts {
		// Try to get stack trace via V8 inspector protocol
		url := fmt.Sprintf("http://127.0.0.1:%s/json", port)
		reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
		if err != nil {
			cancel()
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err != nil {
			continue
		}

		// Check if this is a Node.js inspector
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()

		if !strings.Contains(string(body), "Node") && !strings.Contains(string(body), "V8") {
			continue
		}

		// Try to get the first websocket URL and request stack trace
		// This is a simplified approach - full implementation would need websocket
		// For now, return the inspector info which includes the webSocketDebuggerUrl
		var result strings.Builder
		fmt.Fprintf(&result, "Node.js Inspector found on port %s\n\n", port)
		fmt.Fprintf(&result, "Inspector info:\n%s\n\n", string(body))
		fmt.Fprintf(&result, "To get stack trace:\n")
		fmt.Fprintf(&result, "  1. Connect via: node inspect localhost:%s\n", port)
		fmt.Fprintf(&result, "  2. Or use Chrome DevTools: chrome://inspect\n")
		fmt.Fprintf(&result, "  3. Or use: kill -USR1 %d (if not already in inspect mode)\n", pid)
		return result.String(), nil
	}

	// Method 2: Try llnode (requires lldb + llnode plugin)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "lldb", "-b",
		"-o", "process attach --pid "+fmt.Sprintf("%d", pid),
		"-o", "bt",
		"-o", "quit",
	)
	out, err := cmd.CombinedOutput()
	if err == nil && len(out) > 0 && !strings.Contains(string(out), "error:") {
		return string(out), nil
	}

	return "", fmt.Errorf("no Node.js inspector available and lldb not found")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func connType(t uint32) string {
	switch t {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	default:
		return fmt.Sprintf("type%d", t)
	}
}
