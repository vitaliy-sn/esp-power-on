package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

const powerOnBtnText = "Power On"

var pageTmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ESP Power On{{.DeviceName}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Ccircle cx='16' cy='16' r='14' fill='%2361afef'/%3E%3Cpath d='M16 6v10' stroke='%23282c34' stroke-width='2.5' stroke-linecap='round'/%3E%3Cpath d='M10.5 13.5a7 7 0 1 0 11 0' fill='none' stroke='%23282c34' stroke-width='2.5' stroke-linecap='round'/%3E%3C/svg%3E">
    <style>
        body {
            min-height: 100vh;
            margin: 0;
            padding: 0;
            background: #282c34;
            color: #abb2bf;
            font-family: 'Segoe UI', 'Fira Mono', 'Menlo', 'Consolas', monospace;
        }
        .center-page {
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            width: 100vw;
        }
        .power-btn {
            font-size: 1.1rem;
            text-align: center;
            margin: 0;
            padding: 0 24px;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 48px;
            min-width: 220px;
            background-color: #3e4451;
            color: #61afef;
            border: 1.5px solid #61afef;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(40,44,52,0.5);
            cursor: pointer;
            transition: background 0.2s, box-shadow 0.2s, color 0.2s, border 0.2s;
            font-weight: 500;
            outline: none;
        }
        .power-btn:hover:not(:disabled) {
            background-color: #282c34;
            color: #61afef;
            border: 1.5px solid #abb2bf;
        }
        .power-btn:active:not(:disabled) {
            background-color: #21252b;
            color: #61afef;
            border: 1.5px solid #61afef;
        }
        .power-btn:disabled {
            background-color: #21252b;
            color: #5c6370;
            border: 1.5px solid #5c6370;
            cursor: not-allowed;
            box-shadow: none;
        }
        .custom-snackbar {
            position: fixed;
            top: 32px;
            right: 32px;
            min-width: 240px;
            max-width: 344px;
            padding: 16px 24px;
            border-radius: 8px;
            font-size: 1rem;
            z-index: 9999;
            box-shadow: 0 2px 8px rgba(0,0,0,0.18);
            display: flex;
            align-items: center;
            cursor: pointer;
            user-select: none;
            transition: opacity 0.3s;
            opacity: 0.95;
            word-break: break-word;
        }
        .snackbar-success {
            background-color: rgba(40, 44, 52, 0.95);
            color: #98c379;
            border: 1px solid #98c379;
        }
        .snackbar-error {
            background-color: rgba(40, 44, 52, 0.95);
            color: #e06c75;
            border: 1px solid #e06c75;
        }
    </style>
</head>
<body>
    <div class="center-page">
        <button id="powerBtn" class="power-btn">{{.PowerOnBtnText}}{{.DeviceName}}</button>
        <div id="snackbar" class="custom-snackbar" style="display:none"></div>
    </div>
    <script>
        const btn = document.getElementById('powerBtn');
        const snackbar = document.getElementById('snackbar');
        let cooldown = false;
        btn.addEventListener('click', async function() {
            if (cooldown) return;
            btn.disabled = true;
            btn.innerHTML = "Processing...";
            showSnackbar("", false, true); // hide
            try {
                const resp = await fetch('/poweron', {method: 'POST'});
                if (resp.ok) {
                    showSnackbar("Command sent successfully", true);
                } else {
                    showSnackbar("Failed to send command", false);
                }
            } catch (e) {
                showSnackbar("Network error", false);
            }
            btn.innerHTML = "{{.PowerOnBtnText}}{{.DeviceName}}";
            cooldown = true;
            setTimeout(() => { cooldown = false; btn.disabled = false; }, 5000);
        });
        function showSnackbar(msg, success, hide) {
            if (hide) {
                snackbar.style.display = "none";
                return;
            }
            snackbar.innerText = msg;
            snackbar.className = "custom-snackbar " + (success ? "snackbar-success" : "snackbar-error");
            snackbar.style.display = "flex";
            setTimeout(() => { snackbar.style.display = "none"; }, 4000);
        }
        snackbar.onclick = () => snackbar.style.display = "none";
    </script>
</body>
</html>
`))

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.Info("incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
		next.ServeHTTP(w, r)
		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func main() {
	viper.SetDefault("APP_PORT", "8080")
	viper.AutomaticEnv()

	port := viper.GetString("APP_PORT")
	espAddr := viper.GetString("ESP_ADDRESS")
	deviceName := viper.GetString("DEVICE_NAME")
	if deviceName != "" {
		deviceName = " (" + deviceName + ")"
	}

	if espAddr == "" {
		slog.Error("missing required environment variable", "name", "ESP_ADDRESS")
		os.Exit(1)
	}

	// Print used environment variables and their values
	slog.Info("environment variables",
		"APP_PORT", port,
		"ESP_ADDRESS", espAddr,
		"DEVICE_NAME", deviceName,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			PowerOnBtnText string
			DeviceName     string
		}{
			PowerOnBtnText: powerOnBtnText,
			DeviceName:     deviceName,
		}
		if err := pageTmpl.Execute(w, data); err != nil {
			slog.Error("template error", "err", err)
		}
	})

	mux.HandleFunc("/poweron", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 5 * time.Second}
		url := fmt.Sprintf("http://%s/control?cmd=Pulse,16,1,500", espAddr)
		resp, err := client.Get(url)
		if err != nil {
			slog.Error("poweron request error", "err", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			w.WriteHeader(http.StatusOK)
		} else {
			slog.Error("poweron device returned non-200", "status", resp.Status)
			w.WriteHeader(http.StatusBadGateway)
		}
	})

	addr := fmt.Sprintf(":%s", port)
	slog.Info("server started", "addr", addr, "esp", espAddr)
	if err := http.ListenAndServe(addr, loggingMiddleware(mux)); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
