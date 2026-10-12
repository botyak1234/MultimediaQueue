import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';

export default function AuthModal() {
  const [isLogin, setIsLogin] = useState(true);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const endpoint = isLogin ? '/api/auth/login' : '/api/auth/register';
    try {
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data.error?.message || 'Произошла ошибка при аутентификации');
        return;
      }

      login(data);
    } catch (err) {
      setError('Ошибка соединения с сервером');
    }
  };

  return (
    <div className="auth-wrapper">
      <div className="auth-card">
        <h2 className="auth-title">{isLogin ? 'Вход в Media Planner' : 'Регистрация'}</h2>
        <p className="auth-subtitle">
          {isLogin ? 'Введите логин и пароль' : 'Создайте новый аккаунт'}
        </p>

        {error && <div className="error-banner">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Имя пользователя</label>
            <input
              type="text"
              className="form-input"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              placeholder="alex"
            />
          </div>

          <div className="form-group">
            <label>Пароль</label>
            <input
              type="password"
              className="form-input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="••••••••"
            />
          </div>

          <button type="submit" className="btn-primary" style={{ marginTop: '12px' }}>
            {isLogin ? 'Войти' : 'Зарегистрироваться'}
          </button>
        </form>

        <div className="auth-toggle">
          {isLogin ? 'Еще нет аккаунта?' : 'Уже зарегистрированы?'}
          <span onClick={() => { setIsLogin(!isLogin); setError(''); }}>
            {isLogin ? 'Зарегистрироваться' : 'Войти'}
          </span>
        </div>
      </div>
    </div>
  );
}
