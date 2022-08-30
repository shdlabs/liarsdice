import React, { FC, useContext } from 'react'
import { GameContext } from '../gameContext'
import { CupProps } from '../types/index.d'
import Dice from './dice'

const Cup: FC<CupProps> = (CupProps) => {
  const { player, playerDice } = CupProps
  const { game } = useContext(GameContext)
  const { current_cup, player_order, status } = game

  return player.outs < 3 ? (
    <div data-testid="player__cup" className="player__cup active">
      <Dice
        isPlayerTurn={
          (player_order as string[])[current_cup] === player.account &&
          status === 'playing'
        }
        diceNumber={
          playerDice && status === 'playing' ? playerDice : [0, 0, 0, 0, 0]
        }
      />
    </div>
  ) : (
    <div data-testid="player__cup" className="player__cup"></div>
  )
}

export default Cup
