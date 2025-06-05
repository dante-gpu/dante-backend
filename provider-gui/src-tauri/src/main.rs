#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use tauri::{Manager, Window};
use tauri::api::process::{Command, CommandEvent, CommandChild};
use std::sync::Mutex;

// State to hold the sidecar process handle
struct AppState {
    daemon_child: Mutex<Option<CommandChild>>
}

// Learn more about Tauri commands at https://tauri.app/v1/guides/features/command
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You\'ve been greeted from Rust!", name)
}

#[tauri::command]
async fn start_daemon(window: Window, app_state: tauri::State<'_, AppState>) -> Result<(), String> {
    // Check if a daemon is already running (optional, simple kill for now)
    if let Some(child) = app_state.daemon_child.lock().unwrap().take() {
        child.kill().map_err(|e| format!("Failed to kill previous daemon: {}", e))?;
    }

    let sidecar_name = "provider-daemon"; // Matches externalBin in tauri.conf.json
    
    window.emit("daemon-status", format!("Starting daemon: {} with config.yaml...", sidecar_name)).unwrap_or_default();

    let (mut rx, child) = Command::new_sidecar(sidecar_name)
        .map_err(|e| format!("Failed to create sidecar command (is it in externalBin?): {}", e))?
        .args(["--config", "config.yaml"]) // Instruct daemon to use the config file in its directory
        .spawn()
        .map_err(|e| format!("Failed to spawn sidecar '{}': {}. Ensure it is executable and all dependencies are met.", sidecar_name, e))?;

    // Store the child process handle
    *app_state.daemon_child.lock().unwrap() = Some(child);
    
    window.emit("daemon-status", format!("Daemon {} started.", sidecar_name)).unwrap_or_default();

    // Asynchronously listen to events from the sidecar
    tauri::async_runtime::spawn(async move {
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => {
                    window.emit("daemon-stdout", Some(line)).unwrap_or_default();
                }
                CommandEvent::Stderr(line) => {
                    window.emit("daemon-stderr", Some(line)).unwrap_or_default();
                }
                CommandEvent::Error(message) => {
                    window.emit("daemon-error", Some(format!("Daemon execution error: {}", message))).unwrap_or_default();
                }
                CommandEvent::Terminated(payload) => {
                    let status_message = if let Some(code) = payload.code {
                        format!("Daemon terminated with exit code: {}", code)
                    } else {
                        "Daemon terminated (no exit code or killed)".to_string()
                    };
                    window.emit("daemon-status", Some(status_message)).unwrap_or_default();
                    // The child process handle is managed by AppState and stop_daemon/start_daemon.
                    // No need to explicitly clear here within the async event listener.
                    break; // Exit loop
                }
                _ => {
                    // Other events like Ready, etc. can be handled if needed
                }
            }
        }
    });

    Ok(())
}

#[tauri::command]
async fn stop_daemon(app_state: tauri::State<'_, AppState>) -> Result<(), String> {
    if let Some(child) = app_state.daemon_child.lock().unwrap().take() {
        child.kill().map_err(|e| format!("Failed to kill daemon: {}", e))?;
        Ok(())
    } else {
        Err("Daemon not running or already stopped.".to_string())
    }
}

fn main() {
    tauri::Builder::default()
        .manage(AppState { daemon_child: Mutex::new(None) })
        .invoke_handler(tauri::generate_handler![
            greet,
            start_daemon,
            stop_daemon
        ])
        .setup(|app| {
            let main_window = app.get_window("main").ok_or("Main window not found")?;
            main_window.emit("frontend-ready", Some("Rust backend is ready!")).unwrap_or_default();
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running dante provider gui application");
} 