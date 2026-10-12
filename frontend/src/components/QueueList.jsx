import React, { useEffect, useState } from 'react';

export default function QueueList({ token, onQueueChanged }) {
  const [queue, setQueue] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchQueue = async () => {
    try {
      const res = await fetch('/api/queue', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error('Ошибка загрузки очереди');
      const data = await res.json();
      setQueue(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchQueue();
  }, [token]);

  const updateStatus = async (queueID, status) => {
    try {
      await fetch(`/api/queue/${queueID}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ status }),
      });
      fetchQueue();
      if (onQueueChanged) onQueueChanged();
    } catch (err) {
      alert('Не удалось обновить статус');
    }
  };

  const deleteItem = async (queueID) => {
    try {
      await fetch(`/api/queue/${queueID}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      fetchQueue();
      if (onQueueChanged) onQueueChanged();
    } catch (err) {
      alert('Не удалось удалить из очереди');
    }
  };

  if (loading) return <div>Загрузка очереди...</div>;

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '1.25rem' }}>Очередь просмотра и чтения</h2>
      {error && <div className="error-banner">{error}</div>}

      {queue.length === 0 ? (
        <div style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '40px 0' }}>
          Очередь пуста. Перейдите в "Библиотеку" и добавьте контент в очередь!
        </div>
      ) : (
        <div className="grid-cards">
          {queue.map((q) => {
            const m = q.media;
            if (!m) return null;

            let badgeClass = 'badge-movie';
            let typeLabel = 'Фильм';
            let durationText = `${m.duration_minutes} мин`;

            if (m.type === 'series') {
              badgeClass = 'badge-series';
              typeLabel = 'Сериал';
              durationText = `${m.watched_episodes}/${m.total_episodes} сер. (${m.avg_episode_duration} мин/сер)`;
            } else if (m.type === 'book') {
              badgeClass = 'badge-book';
              typeLabel = 'Книга';
              durationText = `${m.read_pages}/${m.total_pages} стр. (${m.avg_page_read_time} мин/стр)`;
            }

            return (
              <div key={q.id} className="media-card">
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

                <div>
                  <div className="form-group" style={{ marginBottom: '8px' }}>
                    <label style={{ fontSize: '0.75rem' }}>Статус</label>
                    <select
                      className="form-select"
                      style={{ padding: '6px 10px', fontSize: '0.85rem' }}
                      value={q.status}
                      onChange={(e) => updateStatus(q.id, e.target.value)}
                    >
                      <option value="planned">Запланировано</option>
                      <option value="in_progress">В процессе</option>
                      <option value="completed">Завершено</option>
                      <option value="dropped">Отложено</option>
                    </select>
                  </div>

                  <div className="media-actions">
                    <button className="btn-danger" style={{ width: '100%' }} onClick={() => deleteItem(q.id)}>
                      Убрать из очереди
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
