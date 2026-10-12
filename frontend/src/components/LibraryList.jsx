import React, { useEffect, useState } from 'react';

export default function LibraryList({ token, onAddToQueue, onMediaDeleted }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchItems = async () => {
    try {
      const res = await fetch('/api/media', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error('Не удалось загрузить библиотеку');
      const data = await res.json();
      setItems(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchItems();
  }, [token]);

  const handleDelete = async (id) => {
    if (!window.confirm('Удалить этот элемент из библиотеки?')) return;
    try {
      await fetch(`/api/media/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      fetchItems();
      if (onMediaDeleted) onMediaDeleted();
    } catch (err) {
      alert('Ошибка при удалении');
    }
  };

  const handleQueueAdd = async (mediaID) => {
    try {
      const res = await fetch('/api/queue', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ media_id: mediaID, status: 'planned', priority: 1 }),
      });
      if (!res.ok) throw new Error('Ошибка добавления');
      alert('Успешно добавлено в очередь!');
      if (onAddToQueue) onAddToQueue();
    } catch (err) {
      alert('Не удалось добавить в очередь');
    }
  };

  if (loading) return <div>Загрузка библиотеки...</div>;

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '1.25rem' }}>Моя библиотека контента</h2>
      {error && <div className="error-banner">{error}</div>}

      {items.length === 0 ? (
        <div style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '40px 0' }}>
          Библиотека пуста. Нажмите «+ Добавить медиа» вверху!
        </div>
      ) : (
        <div className="grid-cards">
          {items.map((m) => {
            let badgeClass = 'badge-movie';
            let typeLabel = 'Фильм';
            let durationText = `${m.duration_minutes} мин`;

            if (m.type === 'series') {
              badgeClass = 'badge-series';
              typeLabel = 'Сериал';
              durationText = `${m.total_episodes} сер. (${m.avg_episode_duration} мин/сер)`;
            } else if (m.type === 'book') {
              badgeClass = 'badge-book';
              typeLabel = 'Книга';
              durationText = `${m.total_pages} стр. (${m.avg_page_read_time} мин/стр)`;
            }

            return (
              <div key={m.id} className="media-card">
                <div>
                  <span className={`media-badge ${badgeClass}`}>{typeLabel}</span>
                  <div className="media-title">{m.title}</div>
                  <div className="media-meta">
                    {m.release_year > 0 ? `${m.release_year} г. • ` : ''} {durationText}
                  </div>
                  {m.description && (
                    <div style={{ fontSize: '0.85rem', color: '#cbd5e1', marginBottom: '12px' }}>
                      {m.description}
                    </div>
                  )}
                </div>

                <div className="media-actions">
                  <button
                    className="btn-primary"
                    style={{ fontSize: '0.85rem', padding: '6px 12px' }}
                    onClick={() => handleQueueAdd(m.id)}
                  >
                    + В очередь
                  </button>
                  <button
                    className="btn-danger"
                    onClick={() => handleDelete(m.id)}
                  >
                    Удалить
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
