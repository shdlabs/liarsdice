import axios from 'axios'
let fetchedAddress = '0x531130464929826c57BBBF989e44085a02eeB120'
console.log(process.env)
export const getContractAddress = () => fetchedAddress
// export const getContractAddress = () => {
//   axios
//     .get('id.env')
//     .then((res) => {
//       fetchedAddress = res.data
//     })
//     .catch((err) => console.log(err))
//   return fetchedAddress
// }
