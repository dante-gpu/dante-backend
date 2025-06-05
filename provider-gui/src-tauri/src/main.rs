#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::{Serialize, Deserialize};
use tauri::{Manager, State, SystemTrayEvent, Window, AppHandle};
use tauri::api::process::{Command as TauriCommand, CommandEvent, ExitStatus as TauriExitStatus, Child as TauriChild};
use std::sync::Mutex;
use std::process::{Command, Stdio, Child};
use std::io::{BufReader, BufRead, Write};
use std::thread;
use std::time::{SystemTime, UNIX_EPOCH, Duration};
use std::sync::Arc;

#[derive(Clone, Serialize)]
struct LogEntry {
    id: usize,
    message: String,
    timestamp: String,
    log_type: String, // 'status', 'stdout', 'stderr', 'error'
}

#[derive(Serialize, Deserialize, Debug, Clone)]
struct GpuInfo {
    id: String,
    name: String,
    model: String,
    vram_total_mb: u32,
    vram_free_mb: u32,
    utilization_gpu_percent: Option<u32>,
    temperature_c: Option<u32>,
    power_draw_w: Option<u32>,
    is_available_for_rent: bool,
    current_hourly_rate_dgpu: Option<f32>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
struct ProviderSettings {
    default_hourly_rate_dgpu: f32,
    preferred_currency: String,
    min_job_duration_minutes: u32,
    max_concurrent_jobs: u32,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
struct LocalJob {
    id: String,
    name: String,
    status: String, // 'running' | 'completed' | 'failed' | 'queued'
    progress_percent: f32,
    submitted_at: String,
    started_at: Option<String>,
    completed_at: Option<String>,
    estimated_cost_dgpu: Option<f32>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
struct NetworkStatus {
    connection_type: String, // "Ethernet", "WiFi", "Disconnected"
    ip_address: Option<String>,
    upload_speed_mbps: f32,
    download_speed_mbps: f32,
    latency_ms: u32,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
struct FinancialSummary {
    current_balance_dgpu: f32,
    total_earned_dgpu: f32,
    pending_payout_dgpu: f32,
    last_payout_at: Option<String>,
}

struct DaemonState {
    process: Mutex<Option<TauriChild>>,
    log_id_counter: Mutex<usize>,
    status: Mutex<String>, // "offline", "starting", "online", "stopping", "error"
}

impl DaemonState {
    fn new() -> Self {
        DaemonState {
            process: Mutex::new(None),
            log_id_counter: Mutex::new(0),
            status: Mutex::new("offline".to_string()),
        }
    }
}

fn get_timestamp() -> String {
    let now = SystemTime::now();
    // Using a common timestamp format, adjust if App.tsx expects something different
    humantime::format_rfc3339_seconds(now).to_string()
}

fn emit_log_entry<R: tauri::Runtime>(manager: &impl Manager<R>, log_type: &str, message: String) {
    let current_id = {
        let mut counter = manager.try_state::<DaemonState>().unwrap().log_id_counter.lock().unwrap();
        *counter += 1;
        *counter
    };
    let log_payload = LogEntry {
        id: current_id,
        message,
        timestamp: get_timestamp(),
        log_type: log_type.to_string(),
    };

    if let Err(e) = manager.emit_all("daemon_log", log_payload) {
        eprintln!("Failed to emit log event: {}", e); // Fallback log to stderr
    }
}

#[tauri::command]
async fn start_daemon(app_handle: AppHandle, state: State<'_, DaemonState>) -> Result<String, String> {
    let mut status_lock = state.status.lock().unwrap();
    if *status_lock == "online" || *status_lock == "starting" {
        let msg = "Daemon is already online or starting.".to_string();
        emit_log_entry(&app_handle, "status", msg.clone());
        return Ok(msg);
    }
    *status_lock = "starting".to_string();
    emit_log_entry(&app_handle, "status", "Attempting to start provider daemon...".to_string());

    let sidecar_name = "provider-daemon"; // This must match an entry in tauri.conf.json sidecar list or externalBin

    let (mut event_rx, child) = TauriCommand::new_sidecar(sidecar_name)
        .map_err(|e| {
            let err_msg = format!("Failed to create sidecar command '{}'. Ensure it's in tauri.conf.json under externalBin and/or as a sidecar. Error: {}", sidecar_name, e);
            emit_log_entry(&app_handle, "error", err_msg.clone());
            *status_lock = "error".to_string();
            err_msg
        })?
        // .args(&["--daemon-mode"]) // Add any arguments your daemon needs to start in its operational mode
        .spawn()
        .map_err(|e| {
            let err_msg = format!("Failed to spawn sidecar '{}': {}", sidecar_name, e);
            emit_log_entry(&app_handle, "error", err_msg.clone());
            *status_lock = "error".to_string();
            err_msg
        })?;

    let mut process_lock = state.process.lock().unwrap();
    *process_lock = Some(child);
    *status_lock = "online".to_string(); // Set to online once spawn is successful

    emit_log_entry(&app_handle, "status", format!("Daemon process {} started successfully.", sidecar_name));
    
    let app_handle_clone = app_handle.clone();
    let status_mutex_clone = state.status.clone(); 

    tauri::async_runtime::spawn(async move {
        while let Some(event) = event_rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => {
                    emit_log_entry(&app_handle_clone, "stdout", line);
                }
                CommandEvent::Stderr(line) => {
                    emit_log_entry(&app_handle_clone, "stderr", line);
                }
                CommandEvent::Error(message) => {
                    emit_log_entry(&app_handle_clone, "error", format!("Daemon execution error: {}", message));
                    let mut status_guard = status_mutex_clone.lock().unwrap();
                    *status_guard = "error".to_string();
                }
                CommandEvent::Terminated(payload) => {
                    let exit_code_str = payload.code.map_or_else(|| "killed by signal".to_string(), |c| c.to_string());
                    let signal_str = payload.signal.map_or_else(String::new, |s| format!(", signal: {}", s));
                    emit_log_entry(&app_handle_clone, "status", format!("Daemon terminated. Exit code: {}{}", exit_code_str, signal_str));
                    
                    let mut status_guard = status_mutex_clone.lock().unwrap();
                    let previous_status_for_logic = status_guard.clone(); // Clone status before modification
                    
                    // Always set to offline first, then refine to error if needed
                    *status_guard = "offline".to_string(); 

                    if let Some(daemon_state_gaurd) = app_handle_clone.try_state::<DaemonState>() {
                        let mut process_guard = daemon_state_gaurd.process.lock().unwrap();
                         *process_guard = None; // Clear the stored child process
                    } else {
                         emit_log_entry(&app_handle_clone, "error", "Failed to get DaemonState to clear process.".to_string());
                    }

                    if previous_status_for_logic != "stopping" { // If not stopped intentionally
                        if payload.code.is_some() && payload.code != Some(0) {
                             *status_guard = "error".to_string();
                             emit_log_entry(&app_handle_clone, "error", format!("Daemon exited with non-zero status: {}", exit_code_str));
                        } else if payload.code.is_none() { // Killed by signal or other non-exit-code termination
                             *status_guard = "error".to_string();
                             emit_log_entry(&app_handle_clone, "error", "Daemon terminated unexpectedly (e.g. by signal).".to_string());
                        }
                    } else {
                         // If it was stopping, and terminated, it's now offline.
                         emit_log_entry(&app_handle_clone, "status", "Daemon stopped as expected.".to_string());
                    }
                    break; // Exit the event loop once terminated
                }
                CommandEvent::Completed(_payload) => { 
                    // This event is typically for Command::output(), not Command::spawn(). 
                    // It's unlikely to occur here for a long-running daemon.
                    emit_log_entry(&app_handle_clone, "status", "Daemon command marked completed (unexpected for spawned daemon).".to_string());
                }
                _ => { // Catch-all for other events like Running, etc.
                     emit_log_entry(&app_handle_clone, "status", format!("Daemon event: {:?}", event));
                }
            }
        }
        // If the loop exits, it means the event stream ended.
        let mut status_guard = status_mutex_clone.lock().unwrap();
        if *status_guard == "online" || *status_guard == "starting" { 
            *status_guard = "offline".to_string();
            emit_log_entry(&app_handle_clone, "error", "Daemon event stream ended unexpectedly. Marking as offline.".to_string());
        }
    });

    Ok("Daemon started successfully and events are being monitored.".to_string())
}

#[tauri::command]
async fn stop_daemon(app_handle: AppHandle, state: State<'_, DaemonState>) -> Result<String, String> {
    let mut status_lock = state.status.lock().unwrap();
    if *status_lock == "offline" || *status_lock == "stopping" {
        let msg = "Daemon is already offline or stopping.".to_string();
        emit_log_entry(&app_handle, "status", msg.clone());
        return Ok(msg);
    }
    
    let mut process_option_lock = state.process.lock().unwrap();
    if let Some(child_to_kill) = process_option_lock.as_ref() { // Borrow to call kill
        emit_log_entry(&app_handle, "status", "Attempting to stop daemon...".to_string());
        *status_lock = "stopping".to_string(); // Set status before attempting to kill
        drop(status_lock); // Release status_lock before process_option_lock is potentially held longer

        match child_to_kill.kill() {
            Ok(_) => {
                emit_log_entry(&app_handle, "status", "Daemon kill signal sent.".to_string());
                // The CommandEvent::Terminated handler will update the status to "offline"
                // and clear the process from DaemonState.
                Ok("Daemon stop signal sent successfully. Waiting for termination event.".to_string())
            }
            Err(e) => {
                let err_msg = format!("Failed to send kill signal to daemon: {}. Marking as error.", e);
                emit_log_entry(&app_handle, "error", err_msg.clone());
                let mut status_lock_after_fail = state.status.lock().unwrap(); // Re-acquire lock
                *status_lock_after_fail = "error".to_string(); 
                // Also try to clear the process if kill failed, as it might be in an undefined state
                *process_option_lock = None;
                Err(err_msg)
            }
        }
    } else {
        let msg = "No active daemon process found to stop.".to_string();
        emit_log_entry(&app_handle, "status", msg.clone());
        *status_lock = "offline".to_string(); 
        Ok(msg)
    }
}

#[tauri::command]
async fn get_daemon_status(state: State<'_, DaemonState>) -> Result<String, String> {
    Ok(state.status.lock().unwrap().clone())
}

// Helper function to call daemon CLI and parse JSON output
async fn invoke_daemon_cli_json_output<T: for<'de> serde::Deserialize<'de>>(
    app_handle: &tauri::AppHandle,
    command_args: &[&str],
) -> Result<T, String> {
    let sidecar_name = "provider-daemon"; // Matches externalBin if that's the alias for providerd

    // It's good practice to ensure the command name here matches what's in tauri.conf.json's externalBin
    // For example, if externalBin is ["bin/providerd"], sidecar_name might need to reflect that,
    // or tauri::api::process::Command::new() with resolved path might be more robust if not using simple alias.
    // Assuming "provider-daemon" is the direct alias for the executable.
    
    emit_log_entry(app_handle, "status", format!("Invoking daemon: {} with args {:?}", sidecar_name, command_args));

    match tauri::api::process::Command::new_sidecar(sidecar_name)
        .map_err(|e| format!("Sidecar command '{}' not found or misconfigured. Did you add it to tauri.conf.json externalBin/sidecar? Error: {}", sidecar_name, e))?
        .args(command_args)
        .output()
    {
        Ok(output) => {
            if output.status.success() {
                let stdout_str = &output.stdout;
                emit_log_entry(app_handle, "stdout", format!("Daemon response for {:?}: {}", command_args, stdout_str));
                serde_json::from_str(stdout_str)
                    .map_err(|e| {
                        let err_msg = format!("Failed to parse JSON from daemon for {:?}: {}. Output: '{}'", command_args, e, stdout_str);
                        emit_log_entry(app_handle, "error", err_msg.clone());
                        err_msg
                    })
            } else {
                let stderr_str = &output.stderr;
                let stdout_str = &output.stdout;
                let err_msg = format!(
                    "Daemon command {:?} failed with status {:?}: stderr: '{}', stdout: '{}'",
                    command_args, output.status, stderr_str, stdout_str
                );
                emit_log_entry(app_handle, "error", err_msg.clone());
                Err(err_msg)
            }
        }
        Err(e) => {
            let err_msg = format!("Failed to execute daemon command {:?}: {}", command_args, e);
            emit_log_entry(app_handle, "error", err_msg.clone());
            Err(err_msg)
        }
    }
}

// --- New Mock Data Commands ---

#[tauri::command]
async fn get_detected_gpus(app_handle: tauri::AppHandle) -> Result<Vec<GpuInfo>, String> {
    // Real implementation: Call provider-daemon CLI
    // The provider-daemon (Go app) needs to implement a command like:
    // providerd --get-gpus-json
    // This command should print a JSON array of GpuInfo objects to stdout.
    emit_log_entry(&app_handle, "status", "Attempting to fetch GPUs from daemon...".to_string());
    invoke_daemon_cli_json_output::<Vec<GpuInfo>>(&app_handle, &["--get-gpus-json"]).await
}

#[tauri::command]
async fn get_provider_settings(app_handle: tauri::AppHandle) -> Result<ProviderSettings, String> {
    // Real implementation: Call provider-daemon CLI
    // The provider-daemon (Go app) needs to implement a command like:
    // providerd --get-settings-json
    // This command should print a JSON ProviderSettings object to stdout.
    emit_log_entry(&app_handle, "status", "Attempting to fetch provider settings from daemon...".to_string());
    invoke_daemon_cli_json_output::<ProviderSettings>(&app_handle, &["--get-settings-json"]).await
}

#[tauri::command]
async fn update_provider_settings(app_handle: tauri::AppHandle, settings: ProviderSettings) -> Result<ProviderSettings, String> {
    // Real implementation: Call provider-daemon CLI
    // The provider-daemon (Go app) needs to implement a command like:
    // providerd --update-settings-json '{...settings_json...}'
    // This command should save the settings and print the updated (or confirmed) ProviderSettings JSON to stdout.
    emit_log_entry(&app_handle, "status", format!("Attempting to update provider settings via daemon: {:?}", settings));
    let settings_json = serde_json::to_string(&settings)
        .map_err(|e| format!("Failed to serialize settings to JSON: {}", e))?;
    
    invoke_daemon_cli_json_output::<ProviderSettings>(&app_handle, &["--update-settings-json", &settings_json]).await
}

#[tauri::command]
async fn set_gpu_rental_config(app_handle: tauri::AppHandle, gpu_id: String, hourly_rate: f32, available: bool) -> Result<GpuInfo, String> {
    // Real implementation: Call provider-daemon CLI
    // The provider-daemon (Go app) needs to implement a command like:
    // providerd --set-gpu-config-json --gpu-id <gpu_id> --rate <hourly_rate> --available <true|false>
    // This command should update the GPU config and print the updated GpuInfo JSON to stdout.
    emit_log_entry(&app_handle, "status", format!("Attempting to set GPU rental config via daemon: GPU ID {}, Rate {}, Available {}", gpu_id, hourly_rate, available));
    
    invoke_daemon_cli_json_output::<GpuInfo>(&app_handle, &[
        "--set-gpu-config-json",
        "--gpu-id", &gpu_id,
        "--rate", &hourly_rate.to_string(),
        "--available", &available.to_string(),
    ]).await
}


#[tauri::command]
async fn get_local_jobs(app_handle: tauri::AppHandle) -> Result<Vec<LocalJob>, String> {
    // Real implementation: Call provider-daemon CLI
    // The provider-daemon (Go app) needs to implement a command like:
    // providerd --get-local-jobs-json
    // This command should print a JSON array of LocalJob objects to stdout.
    emit_log_entry(&app_handle, "status", "Attempting to fetch local jobs from daemon...".to_string());
    invoke_daemon_cli_json_output::<Vec<LocalJob>>(&app_handle, &["--get-local-jobs-json"]).await
}

#[tauri::command]
async fn get_network_status(app_handle: tauri::AppHandle) -> Result<NetworkStatus, String> {
    // Real implementation: Call provider-daemon CLI or use Rust libraries
    // The provider-daemon (Go app) could implement a command like:
    // providerd --get-network-status-json
    // Or, some parts (like IP) can be fetched using Rust system libraries.
    // Network speeds and latency are more complex and might need dedicated tools/logic in the daemon.
    emit_log_entry(&app_handle, "status", "Attempting to fetch network status from daemon...".to_string());
    invoke_daemon_cli_json_output::<NetworkStatus>(&app_handle, &["--get-network-status-json"]).await
}

#[tauri::command]
async fn get_financial_summary(app_handle: tauri::AppHandle) -> Result<FinancialSummary, String> {
    // Real implementation: Call provider-daemon CLI
    // The provider-daemon (Go app) would use its billing client to get this info, then expose via:
    // providerd --get-financial-summary-json
    // This command should print a JSON FinancialSummary object to stdout.
    emit_log_entry(&app_handle, "status", "Attempting to fetch financial summary from daemon...".to_string());
    invoke_daemon_cli_json_output::<FinancialSummary>(&app_handle, &["--get-financial-summary-json"]).await
}


fn main() {
    let daemon_state = DaemonState::new();

    tauri::Builder::default()
        .manage(daemon_state)
        .invoke_handler(tauri::generate_handler![
            start_daemon, 
            stop_daemon,
            get_daemon_status,
            get_detected_gpus,
            get_provider_settings,
            update_provider_settings,
            set_gpu_rental_config,
            get_local_jobs,
            get_network_status,
            get_financial_summary
        ])
        .setup(|app| {
            emit_log_entry(app, "status", "Provider GUI initialized. Daemon is OFFLINE.".to_string());
            
             // Example system tray (optional, customize as needed)
            let tray_handle = app.tray_handle();
            if let Some(tray) = tray_handle {
                 tray.set_tooltip("Dante Provider GUI")?;
            }


            Ok(())
        })
        .on_system_tray_event(|app, event| match event {
            SystemTrayEvent::LeftClick { .. } => {
                let window = app.get_window("main").unwrap();
                window.show().unwrap();
                window.set_focus().unwrap();
            }
            // Add other tray events if needed (e.g., quit, open dashboard)
            _ => {}
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
} 