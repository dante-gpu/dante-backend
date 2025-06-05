import React, { useState, useEffect, useRef } from 'react';
import { invoke } from '@tauri-apps/api/tauri';
import { listen, Event as TauriEvent } from '@tauri-apps/api/event';

interface LogEntry {
  id: number;
  type: 'status' | 'stdout' | 'stderr' | 'error';
  message: string;
  timestamp: string;
}

// --- BEGIN NEW INTERFACES ---
interface GpuInfo {
  id: string;
  name: string;
  model: string;
  vram_total_mb: number;
  vram_free_mb: number;
  utilization_gpu_percent?: number;
  temperature_c?: number;
  power_draw_w?: number;
  is_available_for_rent: boolean;
  current_hourly_rate_dgpu: number | null;
}

interface ProviderSettings {
  default_hourly_rate_dgpu: number;
  preferred_currency: string; // e.g., "USD", "EUR", "DCORE"
  min_job_duration_minutes: number;
  max_concurrent_jobs: number;
}

interface LocalJob {
  id: string;
  name: string;
  status: 'running' | 'completed' | 'failed' | 'queued';
  progress_percent: number;
  started_at: string;
  estimated_time_remaining_seconds?: number;
}

interface NetworkInfo {
  connection_status: 'connected' | 'disconnected' | 'connecting';
  ip_address?: string;
  upload_speed_mbps?: number;
  download_speed_mbps?: number;
}

interface FinancialSummary {
  wallet_balance_dgpu: number;
  total_earned_dgpu: number;
  pending_payout_dgpu: number;
  last_payout_at?: string;
}
// --- END NEW INTERFACES ---

let logIdCounter = 0;

function App() {
  const [daemonStatus, setDaemonStatus] = useState<string>('OFFLINE');
  const [daemonActive, setDaemonActive] = useState<boolean>(false);
  const [daemonError, setDaemonError] = useState<string | null>(null);
  const [daemonLogs, setDaemonLogs] = useState<LogEntry[]>([]);
  const logsEndRef = useRef<null | HTMLDivElement>(null);

  // --- BEGIN NEW STATE VARIABLES ---
  const [gpus, setGpus] = useState<GpuInfo[]>([]);
  const [providerSettings, setProviderSettings] = useState<ProviderSettings | null>(null);
  const [localJobs, setLocalJobs] = useState<LocalJob[]>([]);
  const [networkStatus, setNetworkStatus] = useState<NetworkInfo | null>(null);
  const [financialSummary, setFinancialSummary] = useState<FinancialSummary | null>(null);
  
  const [selectedGpu, setSelectedGpu] = useState<GpuInfo | null>(null);
  const [gpuRentalModalOpen, setGpuRentalModalOpen] = useState<boolean>(false);
  const [newRentalRate, setNewRentalRate] = useState<string>("");
  // --- END NEW STATE VARIABLES ---

  const addLog = (type: LogEntry['type'], message: string) => {
    const newLog: LogEntry = {
      id: logIdCounter++,
      type,
      message,
      timestamp: new Date().toISOString(),
    };
    setDaemonLogs((prevLogs) => [...prevLogs.slice(-200), newLog]);
  };

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [daemonLogs]);

  useEffect(() => {
    const unlisteners: Array<() => void> = [];

    const setupListener = async <T,>(eventName: string, handler: (event: TauriEvent<T>) => void) => {
      const unlisten = await listen<T>(eventName, handler);
      unlisteners.push(unlisten);
    };

    setupListener<string>('daemon-status', (event) => {
      const statusPayload = event.payload;
      setDaemonStatus(statusPayload);
      setDaemonError(null);
      if (statusPayload.toLowerCase().includes('running') || statusPayload.toLowerCase().includes('started')) {
        setDaemonActive(true);
      } else if (statusPayload.toLowerCase().includes('stopped') || statusPayload.toLowerCase().includes('killed') || statusPayload.toLowerCase().includes('not started') || statusPayload.toLowerCase().includes('offline')) {
        setDaemonActive(false);
      }
      addLog('status', statusPayload);
    });

    setupListener<string>('daemon-stdout', (event) => {
      addLog('stdout', event.payload);
    });

    setupListener<string>('daemon-stderr', (event) => {
      addLog('stderr', event.payload);
    });

    setupListener<string>('daemon-error', (event) => {
      const errorMsg = `Error: ${event.payload}`;
      setDaemonStatus(errorMsg);
      setDaemonError(event.payload); 
      setDaemonActive(false);
      addLog('error', errorMsg);
    });
    
    addLog('status', 'Provider GUI initialized. Daemon is OFFLINE.');

    // --- BEGIN FETCHING INITIAL DATA ---
    const fetchInitialData = async () => {
      if (daemonActive) {
        try {
          const detectedGpus = await invoke<GpuInfo[]>('get_detected_gpus');
          setGpus(detectedGpus);
          addLog('status', `Fetched ${detectedGpus.length} GPUs.`);
        } catch (err) {
          addLog('error', `Failed to get GPUs: ${err}`);
        }

        try {
          const settings = await invoke<ProviderSettings>('get_provider_settings');
          setProviderSettings(settings);
          addLog('status', 'Fetched provider settings.');
        } catch (err) {
          addLog('error', `Failed to get provider settings: ${err}`);
        }
        
        try {
          const jobs = await invoke<LocalJob[]>('get_local_jobs');
          setLocalJobs(jobs);
          addLog('status', `Fetched ${jobs.length} local jobs.`);
        } catch (err) {
          addLog('error', `Failed to get local jobs: ${err}`);
        }

        try {
          const netInfo = await invoke<NetworkInfo>('get_network_status');
          setNetworkStatus(netInfo);
          addLog('status', 'Fetched network status.');
        } catch (err) {
          addLog('error', `Failed to get network status: ${err}`);
        }

        try {
          const finSummary = await invoke<FinancialSummary>('get_financial_summary');
          setFinancialSummary(finSummary);
          addLog('status', 'Fetched financial summary.');
        } catch (err) {
          addLog('error', `Failed to get financial summary: ${err}`);
        }
      }
    };

    if (daemonActive) {
      fetchInitialData();
    }
    // TODO: Consider adding listeners for real-time updates to this data if backend supports it.
    // e.g., await listen('gpu-update', (event) => { setGpus(event.payload); });
     // --- END FETCHING INITIAL DATA ---

    return () => {
      unlisteners.forEach(unlisten => unlisten());
    };
  }, [daemonActive]); // Re-fetch data when daemon becomes active

  const handleStartDaemon = async () => {
    addLog('status', 'Attempting to start daemon...');
    setDaemonError(null);
    try {
      await invoke('start_daemon');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const UImessage = `Failed to send start command: ${errorMessage}`;
      setDaemonStatus(UImessage);
      setDaemonError(errorMessage);
      setDaemonActive(false);
      addLog('error', UImessage);
    }
  };

  const handleStopDaemon = async () => {
    addLog('status', 'Attempting to stop daemon...');
    setDaemonError(null);
    try {
      await invoke('stop_daemon');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const UImessage = `Failed to send stop command: ${errorMessage}`;
      setDaemonStatus(UImessage);
      setDaemonError(errorMessage);
      addLog('error', UImessage);
    }
  };

  const getLogClass = (type: LogEntry['type']) => {
    switch (type) {
      case 'stdout': return 'log-stdout';
      case 'stderr': return 'log-stderr';
      case 'error': return 'log-error';
      case 'status': return 'log-status';
      default: return '';
    }
  };

  const getStatusDisplayClass = () => {
    if (daemonError) return 'status-error';
    if (daemonActive) return 'status-on';
    return 'status-off';
  };

  // --- BEGIN HANDLERS FOR NEW FEATURES ---
  const handleOpenGpuRentalModal = (gpu: GpuInfo) => {
    setSelectedGpu(gpu);
    setNewRentalRate(gpu.current_hourly_rate_dgpu?.toString() || providerSettings?.default_hourly_rate_dgpu.toString() || "0");
    setGpuRentalModalOpen(true);
  };

  const handleCloseGpuRentalModal = () => {
    setGpuRentalModalOpen(false);
    setSelectedGpu(null);
    setNewRentalRate("");
  };

  const handleUpdateGpuRental = async () => {
    if (!selectedGpu) return;
    const rate = parseFloat(newRentalRate);
    if (isNaN(rate) || rate < 0) {
      addLog('error', "Invalid rental rate provided.");
      return;
    }
    try {
      await invoke('update_gpu_rental_settings', { 
        gpuId: selectedGpu.id, 
        isAvailable: !selectedGpu.is_available_for_rent, // Toggle availability
        hourlyRate: rate 
      });
      addLog('status', `Successfully updated rental settings for GPU ${selectedGpu.name}. Toggled availability.`);
      // Refresh GPU list
      const updatedGpus = await invoke<GpuInfo[]>('get_detected_gpus');
      setGpus(updatedGpus);
    } catch (err) {
      addLog('error', `Failed to update GPU ${selectedGpu.name} rental settings: ${err}`);
    }
    handleCloseGpuRentalModal();
  };
  
  const handleToggleGpuAvailability = async (gpu: GpuInfo) => {
    try {
      await invoke('update_gpu_rental_settings', { 
        gpuId: gpu.id, 
        isAvailable: !gpu.is_available_for_rent,
        hourlyRate: gpu.current_hourly_rate_dgpu // Keep current rate or use default if null
      });
      addLog('status', `Toggled availability for GPU ${gpu.name}.`);
      const updatedGpus = await invoke<GpuInfo[]>('get_detected_gpus');
      setGpus(updatedGpus);
    } catch (err) {
      addLog('error', `Failed to toggle availability for GPU ${gpu.name}: ${err}`);
    }
  };

  const handleSaveProviderSettings = async () => {
    if (!providerSettings) return;
    try {
      await invoke('save_provider_settings', { settings: providerSettings });
      addLog('status', 'Provider settings saved.');
    } catch (err) {
      addLog('error', `Failed to save provider settings: ${err}`);
    }
  };
  // --- END HANDLERS FOR NEW FEATURES ---

  return (
    <div className="app-container">
      <header className="header">
        <h1>Dante GPU Provider</h1>
        <p>Manage your provider daemon instance and monitor its activity.</p>
      </header>

      <section className="controls-card">
        <div className={`status-display ${getStatusDisplayClass()}`}>
          Daemon Status: {daemonStatus}
        </div>
        <div className="controls">
          <button onClick={handleStartDaemon} disabled={daemonActive}>
            Start Daemon
          </button>
          <button onClick={handleStopDaemon} disabled={!daemonActive}>
            Stop Daemon
          </button>
        </div>
      </section>

      {/* --- BEGIN GPU MANAGEMENT SECTION --- */}
      <section className="card">
        <h2>GPU Management</h2>
        {daemonActive ? (
          gpus.length > 0 ? (
            <div className="gpu-list">
              {gpus.map(gpu => (
                <div key={gpu.id} className="gpu-item card">
                  <h3>{gpu.name} ({gpu.model})</h3>
                  <p>VRAM: {gpu.vram_free_mb}MB Free / {gpu.vram_total_mb}MB Total</p>
                  {gpu.utilization_gpu_percent !== undefined && <p>Util: {gpu.utilization_gpu_percent}%</p>}
                  {gpu.temperature_c !== undefined && <p>Temp: {gpu.temperature_c}°C</p>}
                  {gpu.power_draw_w !== undefined && <p>Power: {gpu.power_draw_w}W</p>}
                  <p>Status: {gpu.is_available_for_rent ? 
                    <span className="status-rentable">Rentable @ {gpu.current_hourly_rate_dgpu} dGPU/hr</span> : 
                    <span className="status-private">Private</span>}
                  </p>
                  <div className="controls">
                    <button onClick={() => handleOpenGpuRentalModal(gpu)}>
                      {gpu.is_available_for_rent ? 'Edit Rental' : 'Make Rentable'}
                    </button>
                     <button onClick={() => handleToggleGpuAvailability(gpu)}>
                      {gpu.is_available_for_rent ? 'Make Private' : 'Make Rentable (No Price Change)'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p>No GPUs detected or data not yet loaded.</p>
          )
        ) : (
          <p>Daemon is offline. Start the daemon to manage GPUs.</p>
        )}
      </section>
      {/* --- END GPU MANAGEMENT SECTION --- */}

      {/* --- BEGIN RENTAL MODAL --- */}
      {gpuRentalModalOpen && selectedGpu && (
        <div className="modal-backdrop">
          <div className="modal-content card">
            <h3>Set Rental Rate for {selectedGpu.name}</h3>
            <p>Current Status: {selectedGpu.is_available_for_rent ? "Rentable" : "Private"}</p>
            <div>
              <label htmlFor="rentalRate">New Hourly Rate (dGPU): </label>
              <input 
                type="number" 
                id="rentalRate" 
                value={newRentalRate}
                onChange={(e) => setNewRentalRate(e.target.value)}
                min="0"
                step="0.01"
              />
            </div>
            <div className="modal-actions">
              <button onClick={handleUpdateGpuRental}>
                {selectedGpu.is_available_for_rent ? 'Update & Keep Rentable' : 'Set Rate & Make Rentable'}
              </button>
              {selectedGpu.is_available_for_rent && (
                 <button onClick={async () => {
                    if (!selectedGpu) return;
                    try {
                      await invoke('update_gpu_rental_settings', { 
                        gpuId: selectedGpu.id, 
                        isAvailable: false, 
                        hourlyRate: selectedGpu.current_hourly_rate_dgpu 
                      });
                      addLog('status', `GPU ${selectedGpu.name} set to Private.`);
                      const updatedGpus = await invoke<GpuInfo[]>('get_detected_gpus');
                      setGpus(updatedGpus);
                    } catch (err) {
                      addLog('error', `Failed to set GPU ${selectedGpu.name} to private: ${err}`);
                    }
                    handleCloseGpuRentalModal();
                  }}>
                  Make Private
                </button>
              )}
              <button onClick={handleCloseGpuRentalModal}>Cancel</button>
            </div>
          </div>
        </div>
      )}
      {/* --- END RENTAL MODAL --- */}
      
      {/* --- BEGIN LOCAL JOB MONITORING SECTION --- */}
      <section className="card">
        <h2>Local Job Monitoring</h2>
        {daemonActive ? (
          localJobs.length > 0 ? (
             <div className="job-list">
              {localJobs.map(job => (
                <div key={job.id} className="job-item card">
                  <h4>{job.name} (ID: {job.id})</h4>
                  <p>Status: <span className={`job-status-${job.status}`}>{job.status}</span></p>
                  <p>Progress: {job.progress_percent}%</p>
                  {/* Basic progress bar */}
                  <div style={{ width: '100%', backgroundColor: '#eee' }}>
                    <div style={{ width: `${job.progress_percent}%`, backgroundColor: 'green', height: '10px' }}></div>
                  </div>
                  <p>Started: {new Date(job.started_at).toLocaleString()}</p>
                  {job.estimated_time_remaining_seconds && <p>ETA: {job.estimated_time_remaining_seconds}s</p>}
                </div>
              ))}
            </div>
          ) : (
            <p>No local jobs active or data not yet loaded.</p>
          )
        ) : (
          <p>Daemon is offline. Start the daemon to see local jobs.</p>
        )}
      </section>
      {/* --- END LOCAL JOB MONITORING SECTION --- */}

      {/* --- BEGIN SYSTEM & FINANCIAL OVERVIEW SECTION --- */}
      <section className="card">
        <h2>System & Financial Overview</h2>
        {daemonActive ? (
          <>
            <div className="overview-item">
              <h4>Network Status</h4>
              {networkStatus ? (
                <>
                  <p>Connection: {networkStatus.connection_status}</p>
                  {networkStatus.ip_address && <p>IP: {networkStatus.ip_address}</p>}
                  {/* Add more network details if available */}
                </>
              ) : <p>Loading network info...</p>}
            </div>
            <div className="overview-item">
              <h4>Financial Summary</h4>
              {financialSummary ? (
                <>
                  <p>Wallet Balance: {financialSummary.wallet_balance_dgpu.toFixed(2)} dGPU</p>
                  <p>Total Earned: {financialSummary.total_earned_dgpu.toFixed(2)} dGPU</p>
                  <p>Pending Payout: {financialSummary.pending_payout_dgpu.toFixed(2)} dGPU</p>
                  {financialSummary.last_payout_at && <p>Last Payout: {new Date(financialSummary.last_payout_at).toLocaleString()}</p>}
                </>
              ) : <p>Loading financial summary...</p>}
            </div>
          </>
        ) : (
          <p>Daemon is offline. Start the daemon to see system and financial overview.</p>
        )}
      </section>
      {/* --- END SYSTEM & FINANCIAL OVERVIEW SECTION --- */}

      {/* --- BEGIN PROVIDER SETTINGS SECTION --- */}
      <section className="card">
        <h2>Provider Settings</h2>
        {daemonActive ? (
          providerSettings ? (
            <div>
              <div>
                <label htmlFor="defaultRate">Default Hourly Rate (dGPU): </label>
                <input 
                  type="number" 
                  id="defaultRate" 
                  value={providerSettings.default_hourly_rate_dgpu}
                  onChange={(e) => setProviderSettings({...providerSettings, default_hourly_rate_dgpu: parseFloat(e.target.value)})}
                  min="0"
                  step="0.01"
                />
              </div>
              {/* Add more settings inputs here based on ProviderSettings interface */}
              <button onClick={handleSaveProviderSettings}>Save Settings</button>
            </div>
          ) : (
            <p>Loading provider settings...</p>
          )
        ) : (
          <p>Daemon is offline. Start the daemon to configure settings.</p>
        )}
      </section>
      {/* --- END PROVIDER SETTINGS SECTION --- */}

      <section className="logs-card">
        <h2 className="logs-header">Daemon Activity Logs</h2>
        <div className="logs-container">
          {daemonLogs.map((log) => (
            <div key={log.id} className={`log-entry ${getLogClass(log.type)}`}>
              <span className="log-timestamp">
                {new Date(log.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
              </span>
              <span className="log-message">{log.message}</span>
            </div>
          ))}
          <div ref={logsEndRef} />
        </div>
      </section>
    </div>
  );
}

export default App; 