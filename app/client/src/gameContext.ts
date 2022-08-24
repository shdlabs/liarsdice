import { createContext } from 'react'
import { claim, game } from './types/index.d'

export const GameContext = createContext({
  game: {
    status: 'gameover',
    last_out: '',
    last_win: '',
    current_player: '',
    current_cup: 0,
    round: 1,
    cups: [],
    player_order: [],
    claims: [] as claim[],
    ante_usd: 0,
  } as game,
  setGame: (() => {}) as React.Dispatch<React.SetStateAction<game>>,
})
