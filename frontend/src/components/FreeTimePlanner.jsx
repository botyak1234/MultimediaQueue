import React, { useState } from 'react';

export default function FreeTimePlanner({ token }) {
  const [minutes, setMinutes] = useState(60);
  const [plan, setPlan] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const quickPresets = [30, 45, 60, 90, 120, 180];

  const handleGeneratePlan = async (e) => {
    if (e) e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch('/api/planner', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ available_minutes: parseInt(minutes) }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data.error?.message || 'Не удалось сформировать план');
        setPlan(null);
        return;
      }

      setPlan(data);
    } catch (err) {
      setError('Ошибка соединения при запросе планировщика');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto' }}>
      <div className="auth-card" style={{ maxWidth: '100%' }}>
        <h2 className="auth-title" style={{ textAlign: 'left' }}>⚡ Планировщик досуга</h2>
        <p className="auth-subtitle" style={{ textAlign: 'left', marginBottom: '16px' }}>
          Укажите, сколько у вас есть свободного времени, и система сама подберет идеальное сочетание фильмов, серий или глав из вашей очереди!
        </p>

        {error && <div className="error-banner">{error}</div>}

        <form onSubmit={handleGeneratePlan}>
          <div className="form-group">
            <label>Доступное свободное время (в минутах):</label>
            <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
              <input
                type="number"
                className="form-input"
                style={{ fontSize: '1.2rem', fontWeight: 'bold' }}
                value={minutes}
                onChange={(e) => setMinutes(e.target.value)}
                min="1"
                required
              />
              <button type="submit" className="btn-primary" style={{ width: '220px' }} disabled={loading}>
                {loading ? 'Считаем...' : 'Сформировать план'}
              </button>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', marginTop: '12px' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)', alignSelf: 'center' }}>Быстрый выбор:</span>
            {quickPresets.map((m) => (
              <button
                key={m}
                type="button"
                className="btn-secondary"
                style={{ padding: '4px 10px', fontSize: '0.85rem' }}
                onClick={() => {
                  setMinutes(m);
                }}
              >
                {m >= 60 ? `${m / 60} ч` : `${m} мин`}
              </button>
            ))}
          </div>
        </form>

        {plan && (
          <div className="planner-box">
            <h3 style={{ fontSize: '1.2rem', marginBottom: '8px', color: '#818cf8' }}>
              🎯 Рекомендуемый план просмотра/чтения
            </h3>
            <div style={{ fontSize: '0.95rem', color: 'var(--text-muted)', marginBottom: '16px' }}>
              Задействовано: <strong style={{ color: 'white' }}>{plan.used_minutes} мин</strong> из {plan.available_minutes} мин
              {plan.remaining_minutes > 0 ? ` (Остаток: ${plan.remaining_minutes} мин)` : ''}
            </div>

            {plan.actions.length === 0 ? (
              <div style={{ color: '#fca5a5', padding: '12px 0' }}>
                В вашей очереди нет подходящих пунктов, которые поместились бы в {plan.available_minutes} минут. Попробуйте увеличить время или добавить более короткие фильмы/серии!
              </div>
            ) : (
              <div>
                {plan.actions.map((act, idx) => (
                  <div key={idx} className="action-item">
                    <div style={{ fontWeight: '600', fontSize: '1rem', color: 'white' }}>
                      {idx + 1}. {act.description}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
