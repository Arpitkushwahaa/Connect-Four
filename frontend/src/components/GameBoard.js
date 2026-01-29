import React from 'react';
import './GameBoard.css';

const ROWS = 6;
const COLUMNS = 7;

function GameBoard({ game, onColumnClick, canPlay }) {
  const isWinningCell = (row, col) => {
    if (!game.winningLine) return false;
    return game.winningLine.some(([r, c]) => r === row && c === col);
  };

  const getCellClass = (row, col) => {
    const value = game.board[row][col];
    let className = 'cell';
    
    if (value === 1) className += ' player1';
    else if (value === 2) className += ' player2';
    
    if (isWinningCell(row, col)) {
      className += ' winning';
    }
    
    return className;
  };

  const handleColumnClick = (col) => {
    if (!canPlay) return;
    
    // Check if column is full
    if (game.board[0][col] !== 0) return;
    
    onColumnClick(col);
  };

  const isColumnFull = (col) => {
    return game.board[0][col] !== 0;
  };

  return (
    <div className="game-board">
      <div className="board-container">
        {/* Column click indicators */}
        <div className="column-indicators">
          {Array.from({ length: COLUMNS }).map((_, col) => (
            <div
              key={col}
              className={`column-indicator ${canPlay && !isColumnFull(col) ? 'active' : ''} ${isColumnFull(col) ? 'full' : ''}`}
              onClick={() => handleColumnClick(col)}
            >
              {canPlay && !isColumnFull(col) && (
                <div className="drop-arrow">▼</div>
              )}
            </div>
          ))}
        </div>

        {/* Game board */}
        <div className="board">
          {Array.from({ length: ROWS }).map((_, row) => (
            <div key={row} className="row">
              {Array.from({ length: COLUMNS }).map((_, col) => (
                <div
                  key={`${row}-${col}`}
                  className={getCellClass(row, col)}
                >
                  <div className="disc-slot">
                    {game.board[row][col] !== 0 && (
                      <div className={`disc ${game.board[row][col] === 1 ? 'red' : 'yellow'}`} />
                    )}
                  </div>
                </div>
              ))}
            </div>
          ))}
        </div>

        {/* Board base */}
        <div className="board-base" />
      </div>
    </div>
  );
}

export default GameBoard;
