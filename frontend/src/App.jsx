import React, { useState } from 'react';
import { useAuth } from './context/AuthContext';
import AuthModal from './components/AuthModal';
import LibraryList from './components/LibraryList';
import QueueList from './components/QueueList';
import FreeTimePlanner from './components/FreeTimePlanner';
import AddMediaModal from './components/AddMediaModal';

export default function App() {
  const { user, token, logout, loading } = useAuth();
  const [activeTab, setActiveTab] = useState('queue'); // 'queue', 'library', 'planner'
  const [showAddModal, setShowAddModal] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  if (loading) {
    return <div style={{ textAlign: 'center', marginTop: '100px' }}>Загрузка...</div>;
  }

  if (!user) {
    return <AuthModal />;
  }

  const triggerRefresh = () => setRefreshKey((prev) => prev + 1);

  return (
    <div className="app-container">
      <header className="header">
        <div className="brand">
          <span>🎬</span>
          <span>Media Planner</span>
        </div>

        <nav className="nav-tabs">
          <button
            className={`tab-btn ${activeTab === 'queue' ? 'active' : ''}`}
            onClick={() => setActiveTab('queue')}
          >
            📋 Очередь
          </button>
          <button
            className={`tab-btn ${activeTab === 'planner' ? 'active' : ''}`}
            onClick={() => setActiveTab('planner')}
          >
            ⚡ Планировщик
          </button>
          <button
            className={`tab-btn ${activeTab === 'library' ? 'active' : ''}`}
            onClick={() => setActiveTab('library')}
          >
            📚 Библиотека
          </button>
        </nav>

        <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
          <button className="btn-secondary" onClick={() => setShowAddModal(true)}>
            + Добавить медиа
          </button>
          <span style={{ fontSize: '0.9rem', color: 'var(--text-muted)' }}>
            👤 {user.username}
          </span>
          <button className="btn-danger" onClick={logout}>
            Выйти
          </button>
        </div>
      </header>

      <main>
        {activeTab === 'queue' && <QueueList key={refreshKey} token={token} onQueueChanged={triggerRefresh} />}
        {activeTab === 'planner' && <FreeTimePlanner token={token} />}
        {activeTab === 'library' && (
          <LibraryList
            key={refreshKey}
            token={token}
            onAddToQueue={triggerRefresh}
            onMediaDeleted={triggerRefresh}
          />
        )}
      </main>

      {showAddModal && (
        <AddMediaModal
          token={token}
          onClose={() => setShowAddModal(false)}
          onAdded={() => triggerRefresh()}
        />
      )}
    </div>
  );
}
