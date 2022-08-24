import { game } from '../types/index.d'

const assureGameType = (data: game): game => {
  let newGame = data
  newGame = newGame.claims ? newGame : { ...newGame, claims: [] }
  newGame = newGame.cups ? newGame : { ...newGame, cups: [] }
  newGame = newGame.player_order ? newGame : { ...newGame, player_order: [] }
  return newGame
}
export default assureGameType
