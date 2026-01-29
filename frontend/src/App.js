import React, { useState, useEffect, useRef } from 'react';
import './App.css';
import GameBoard from './components/GameBoard';
import Leaderboard from './components/Leaderboard';

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

function App() {
  const [username, setUsername] = useState('');
  const [gameState, setGameState] = useState('idle'); // idle, waiting, playing, finished
  const [game, setGame] = useState(null);
  const [yourPlayerId, setYourPlayerId] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [showLeaderboard, setShowLeaderboard] = useState(false);
  
  const wsRef = useRef(null);

  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const connectWebSocket = () => {
    const ws = new WebSocket(WS_URL);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      setError('');
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleMessage(data);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setError('Connection error. Please try again.');
    };

    ws.onclose = () => {
      console.log('WebSocket closed');
      if (gameState === 'playing' || gameState === 'waiting') {
        setError('Connection lost. Attempting to reconnect...');
        setTimeout(() => {
          if (game) {
            reconnect();
          }
        }, 2000);
      }
    };

    wsRef.current = ws;
  };

  const handleMessage = (data) => {
    console.log('Received message:', data);

    switch (data.type) {
      case 'game_start':
        setGameState('playing');
        setGame(data.payload.game);
        setYourPlayerId(data.payload.yourPlayerId);
        setMessage(`Game started! You are Player ${data.payload.yourPlayerId === data.payload.game.player1.id ? '1 (Red)' : '2 (Yellow)'}`);
        break;

      case 'game_update':
        setGame(data.payload.game);
        if (data.payload.message) {
          setMessage(data.payload.message);
        }
        break;

      case 'game_over':
        setGame(data.payload.game);
        setGameState('finished');
        setMessage(data.payload.message);
        break;

      case 'error':
        setError(data.payload.message);
        setTimeout(() => setError(''), 5000);
        break;

      case 'invalid_move':
        setError(data.payload.message);
        setTimeout(() => setError(''), 3000);
        break;

      case 'opponent_left':
        setMessage(data.payload.message);
        break;

      default:
        console.log('Unknown message type:', data.type);
    }
  };

  const joinQueue = () => {
    if (!username.trim()) {
      setError('Please enter a username');
      return;
    }

    connectWebSocket();
    
    // Wait for connection to open
    const checkConnection = setInterval(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        clearInterval(checkConnection);
        
        wsRef.current.send(JSON.stringify({
          type: 'join_queue',
          payload: {
            username: username.trim()
          }
        }));

        setGameState('waiting');
        setMessage('Waiting for opponent...');
      }
    }, 100);
  };

  const reconnect = () => {
    if (!game || !username) return;

    connectWebSocket();

    const checkConnection = setInterval(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        clearInterval(checkConnection);
        
        wsRef.current.send(JSON.stringify({
          type: 'reconnect',
          payload: {
            username: username.trim(),
            gameId: game.id
          }
        }));
      }
    }, 100);
  };

  const makeMove = (column) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setError('Not connected to server');
      return;
    }

    if (gameState !== 'playing') {
      return;
    }

    wsRef.current.send(JSON.stringify({
      type: 'move',
      payload: {
        column: column
      }
    }));
  };

  const playAgain = () => {
    setGameState('idle');
    setGame(null);
    setYourPlayerId('');
    setMessage('');
    setError('');
    if (wsRef.current) {
      wsRef.current.close();
    }
  };

  const isYourTurn = () => {
    if (!game || !yourPlayerId || gameState !== 'playing') return false;
    
    const playerNum = yourPlayerId === game.player1.id ? 1 : 2;
    return game.currentTurn === playerNum;
  };

  const getPlayerColor = (playerNum) => {
    return playerNum === 1 ? 'Red' : 'Yellow';
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>🎮 4 in a Row</h1>
        <button 
          className="leaderboard-toggle"
          onClick={() => setShowLeaderboard(!showLeaderboard)}
        >
          {showLeaderboard ? 'Hide Leaderboard' : 'Show Leaderboard'}
        </button>
      </header>

      {showLeaderboard && (
        <div className="leaderboard-container">
          <Leaderboard />
        </div>
      )}

      <main className="App-main">
        {error && (
          <div className="error-message">
            {error}
          </div>
        )}

        {message && (
          <div className="info-message">
            {message}
          </div>
        )}

        {gameState === 'idle' && (
          <div className="join-container">
            <h2>Enter Your Name to Play</h2>
            <input
              type="text"
              placeholder="Your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && joinQueue()}
              className="username-input"
              maxLength={20}
            />
            <button onClick={joinQueue} className="join-button">
              Join Game
            </button>
            <p className="info-text">
              You'll be matched with another player or a bot if no one is available
            </p>
          </div>
        )}

        {gameState === 'waiting' && (
          <div className="waiting-container">
            <div className="loader"></div>
            <h2>Waiting for opponent...</h2>
            <p>A bot will join if no player is found within 10 seconds</p>
          </div>
        )}

        {(gameState === 'playing' || gameState === 'finished') && game && (
          <div className="game-container">
            <div className="game-info">
              <div className="players-info">
                <div className={`player-card ${yourPlayerId === game.player1.id ? 'you' : ''}`}>
                  <span className="player-disc red"></span>
                  <span className="player-name">
                    {game.player1.username}
                    {game.player1.isBot && ' 🤖'}
                    {yourPlayerId === game.player1.id && ' (You)'}
                  </span>
                </div>
                <div className="vs">VS</div>
                <div className={`player-card ${yourPlayerId === game.player2?.id ? 'you' : ''}`}>
                  <span className="player-disc yellow"></span>
                  <span className="player-name">
                    {game.player2?.username}
                    {game.player2?.isBot && ' 🤖'}
                    {yourPlayerId === game.player2?.id && ' (You)'}
                  </span>
                </div>
              </div>

              {gameState === 'playing' && (
                <div className={`turn-indicator ${isYourTurn() ? 'your-turn' : ''}`}>
                  {isYourTurn() ? (
                    <strong>Your Turn! ({getPlayerColor(game.currentTurn)})</strong>
                  ) : (
                    <span>Opponent's Turn ({getPlayerColor(game.currentTurn)})</span>
                  )}
                </div>
              )}

              {gameState === 'finished' && (
                <div className="game-result">
                  <h2>
                    {game.winner 
                      ? `${game.winner.username} Wins! 🎉` 
                      : "It's a Draw! 🤝"}
                  </h2>
                  <button onClick={playAgain} className="play-again-button">
                    Play Again
                  </button>
                </div>
              )}
            </div>

            <GameBoard 
              game={game} 
              onColumnClick={makeMove}
              canPlay={isYourTurn()}
            />
          </div>
        )}
      </main>

      <footer className="App-footer">
        <p>Built with React & GoLang | Real-time multiplayer with WebSockets</p>
      </footer>
    </div>
  );
}

export default App;
