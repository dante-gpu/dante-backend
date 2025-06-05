import React, { useState, useEffect, useRef } from 'react';
import { invoke } from '@tauri-apps/api/tauri';
import { listen, Event as TauriEvent } from '@tauri-apps/api/event';

interface LogEntry {
  id: number;
  type: 'status' | 'stdout' | 'stderr' | 'error';
  message: string;
  timestamp: string;
}

let logIdCounter = 0;

function App() {
  const [daemonStatus, setDaemonStatus] = useState<string>('OFFLINE');
  const [daemonActive, setDaemonActive] = useState<boolean>(false);
  const [daemonError, setDaemonError] = useState<string | null>(null);
  const [daemonLogs, setDaemonLogs] = useState<LogEntry[]>([]);
  const logsEndRef = useRef<null | HTMLDivElement>(null);

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

    return () => {
      unlisteners.forEach(unlisten => unlisten());
    };
  }, []);

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