// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import "./error.sol";

contract LiarsDice {

    // game holds all game related data.
    struct game {
        uint created_at;
        bool finished;
        uint256 pot;
    }

    // Owner represents the address who deployed the contract.
    address public Owner;

    // playerbalance represents the amount of coins a player have available.
    mapping (address => uint) public playerbalance;

    // Games represents a list of all games.
    mapping (string => game) public games;

    // EventLog provides support for external logging.
    event EventLog(string value);

    // EventPlaceAnte is an event to indicate a bet was performed.
    event EventPlaceAnte(address player, string uuid, uint amount);

    // EventNewGame is an event to indicate a new game was created.
    event EventNewGame(string uuid);

    // onlyOwner can be used to restrict access to a function for only the owner.
    modifier onlyOwner {
        if (msg.sender != Owner) revert();
        _;
    }

    // constructor is called when the contract is deployed.
    constructor() {
        Owner = msg.sender;
    }

    // NewGame creates a new game with the given uuid and default values.
    function NewGame(string memory uuid) public {
        games[uuid] = game(block.timestamp, false, 0);
        emit EventNewGame(uuid);
    }

    // PlaceAnte adds the amount to the game pot and removes from player balance.
    function PlaceAnte(string memory uuid, uint256 amount) payable public {
        address player = msg.sender;
        uint gas = gasleft();
        emit EventLog(string.concat(Error.Itoa(msg.value), Error.Itoa(gas)));
        
        // TODO: check the minimum value against the amount.

        // Check if game is finshed.
        if (games[uuid].finished) {
            revert(string.concat("game ", uuid, " is not available anymore"));
        }

        games[uuid].pot += amount;

        emit EventLog(string.concat("player ", Error.Addrtoa(player), " placed a bet of ", Error.Itoa(amount), " LDC on game ", uuid));
        emit EventLog(string.concat("current game pot ", Error.Itoa(games[uuid].pot)));
        emit EventPlaceAnte(player, uuid, amount);
    }

    // GameEnd transfers the game pot amount to the player and finish the game.
    function GameEnd(address player, string memory uuid) public {
        games[uuid].finished = true;

        // TODO: Transfer the pot value to the player's address.

        emit EventLog(string.concat("game ", uuid, " is over with a pot of ", Error.Itoa(games[uuid].pot), " LDC. The winner is ", Error.Addrtoa(player)));
    }

    // GameAnte returns the game pot amount.
    function GameAnte(string memory uuid) onlyOwner public returns (uint) {
        emit EventLog(string.concat("game ", uuid, " has a pot of ", Error.Itoa(games[uuid].pot), " LDCs"));
        return games[uuid].pot;
    }

    //===========================================================================

    // [For testing purposes]
    // deposits the given amount to the player balance.
    function deposit(address player, uint256 amount) public {
        playerbalance[player] += amount;
    }

    // [For testing purposes]
    // withdraws the given amount to the player balance.
    function withdraw(address player, uint256 amount) public {
        playerbalance[player] -= amount;
    }
}