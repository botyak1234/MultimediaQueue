import React, { useState } from 'react';

export default function AddMediaModal({ token, onClose, onAdded }) {
  const [type, setType] = useState('movie');
  const [title, setTitle] = useState('');
  const [releaseYear, setReleaseYear] = useState(new Date().getFullYear());
  const [durationMinutes, setDurationMinutes] = useState(120);
  const [totalEpisodes, setTotalEpisodes] = useState(10);
  const [avgEpisodeDuration, setAvgEpisodeDuration] = useState(45);
  const [totalPages, setTotalPages] = useState(300);
  const [avgPageReadTime, setAvgPageReadTime] = useState(2);
  const [description, setDescription] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const payload = {
      type,
      title,
      release_year: parseInt(releaseYear) || 0,
      description,
    };

    if (type === 'movie') {
      payload.duration_minutes = parseInt(durationMinutes) || 0;
    } else if (type === 'series') {
      payload.total_episodes = parseInt(totalEpisodes) || 0;
      payload.avg_episode_duration = parseInt(avgEpisodeDuration) || 0;
    } else if (type === 'book') {
      payload.total_pages = parseInt(totalPages) || 0;
      payload.avg_page_read_time = parseInt(avgPageReadTime) || 2;
    }

    try {
      const res = await fetch('/api/media', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data.error?.message || 'Ошибка создания элемента');
        return;
      }

      onAdded(data);
      onClose();
    } catch (err) {
      setError('Ошибка соединения');
    }
  };

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      backgroundColor: 'rgba(0,0,0,0.7)', display: 'flex',
      alignItems: 'center', justifyContent: 'center', zIndex: 1000
    }}>
      <div className="auth-card" style={{ maxWidth: '500px' }}>
        <h2 className="auth-title">Добавить контент</h2>
        {error && <div className="error-banner">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Тип медиа</label>
            <select className="form-select" value={type} onChange={(e) => setType(e.target.value)}>
              <option value="movie">Фильм</option>
              <option value="series">Сериал</option>
              <option value="book">Книга</option>
            </select>
          </div>

          <div className="form-group">
            <label>Название</label>
            <input
              type="text"
              className="form-input"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              placeholder="Например: Начало, Дюна, Игра Престолов"
            />
          </div>

          <div className="form-group">
            <label>Год выпуска</label>
            <input
              type="number"
              className="form-input"
              value={releaseYear}
              onChange={(e) => setReleaseYear(e.target.value)}
            />
          </div>

          {type === 'movie' && (
            <div className="form-group">
              <label>Продолжительность (минут)</label>
              <input
                type="number"
                className="form-input"
                value={durationMinutes}
                onChange={(e) => setDurationMinutes(e.target.value)}
                required
              />
            </div>
          )}

          {type === 'series' && (
            <div className="grid-2">
              <div className="form-group">
                <label>Всего серий</label>
                <input
                  type="number"
                  className="form-input"
                  value={totalEpisodes}
                  onChange={(e) => setTotalEpisodes(e.target.value)}
                  required
                />
              </div>
              <div className="form-group">
                <label>Сред. время серии (мин)</label>
                <input
                  type="number"
                  className="form-input"
                  value={avgEpisodeDuration}
                  onChange={(e) => setAvgEpisodeDuration(e.target.value)}
                  required
                />
              </div>
            </div>
          )}

          {type === 'book' && (
            <div className="grid-2">
              <div className="form-group">
                <label>Всего страниц</label>
                <input
                  type="number"
                  className="form-input"
                  value={totalPages}
                  onChange={(e) => setTotalPages(e.target.value)}
                  required
                />
              </div>
              <div className="form-group">
                <label>Мин. на страницу</label>
                <input
                  type="number"
                  className="form-input"
                  value={avgPageReadTime}
                  onChange={(e) => setAvgPageReadTime(e.target.value)}
                  required
                />
              </div>
            </div>
          )}

          <div className="form-group">
            <label>Описание</label>
            <textarea
              className="form-textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows="2"
            />
          </div>

          <div style={{ display: 'flex', gap: '8px', marginTop: '16px' }}>
            <button type="submit" className="btn-primary">Сохранить</button>
            <button type="button" className="btn-secondary" onClick={onClose}>Отмена</button>
          </div>
        </form>
      </div>
    </div>
  );
}
