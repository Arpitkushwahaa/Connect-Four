import React, { useState, useEffect } from 'react';
import './Leaderboard.css';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

function Leaderboard() {
  const [leaderboard, setLeaderboard] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchLeaderboard();
    const interval = setInterval(fetchLeaderboard, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchLeaderboard = async () => {
    try {
      const response = await fetch(`${API_URL}/api/leaderboard`);
      if (!response.ok) {
        throw new Error('Failed to fetch leaderboard');
      }
      const data = await response.json();
      setLeaderboard(data || []);
      setLoading(false);
      setError('');
    } catch (err) {
      setError('Failed to load leaderboard');
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="leaderboard">
        <h2>🏆 Leaderboard</h2>
        <div className="loader-small"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="leaderboard">
        <h2>🏆 Leaderboard</h2>
        <p className="error">{error}</p>
      </div>
    );
  }

  return (
    <div className="leaderboard">
      <h2>🏆 Leaderboard</h2>
      {leaderboard.length === 0 ? (
        <p className="no-data">No games played yet. Be the first!</p>
      ) : (
        <table className="leaderboard-table">
          <thead>
            <tr>
              <th>Rank</th>
              <th>Player</th>
              <th>Wins</th>
              <th>Losses</th>
              <th>Draws</th>
              <th>Win Rate</th>
            </tr>
          </thead>
          <tbody>
            {leaderboard.map((entry, index) => {
              const totalGames = entry.wins + entry.losses + entry.draws;
              const winRate = totalGames > 0 
                ? ((entry.wins / totalGames) * 100).toFixed(1) 
                : '0.0';
              
              return (
                <tr key={entry.username} className={index < 3 ? `top-${index + 1}` : ''}>
                  <td className="rank">
                    {index === 0 && '🥇'}
                    {index === 1 && '🥈'}
                    {index === 2 && '🥉'}
                    {index > 2 && (index + 1)}
                  </td>
                  <td className="username">{entry.username}</td>
                  <td className="wins">{entry.wins}</td>
                  <td className="losses">{entry.losses}</td>
                  <td className="draws">{entry.draws}</td>
                  <td className="win-rate">{winRate}%</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default Leaderboard;
