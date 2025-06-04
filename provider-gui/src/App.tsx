import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/tauri';

function App() {
  const [greeting, setGreeting] = useState('');

  useEffect(() => {
    // Example of invoking a command from the Rust backend
    invoke('greet', { name: 'Provider GUI' })
      .then((response) => setGreeting(response as string))
      .catch(console.error);
  }, []);

  return (
    <div className="container">
      <h1>Welcome to Dante GPU Provider Dashboard</h1>
      <p>This is where you will manage your GPU contributions.</p>
      <p>Message from backend: {greeting}</p>
      {/* Further UI components will go here */}
    </div>
  );
}

export default App; 