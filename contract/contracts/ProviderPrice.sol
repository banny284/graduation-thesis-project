// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

// import openzeppelin contracts
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
// import openzeppelin contracts initializer
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";


// initialize contract
contract PriceProvider is Initializable, OwnableUpgradeable {
    // set price

    struct BTC_USDT_PRICE {
        uint256 price;
        uint256 timestamp;
    }

    // set mapping
    //  round -> price
    mapping(uint256 => BTC_USDT_PRICE)  btc_usdt_price;

    BTC_USDT_PRICE private btc_usdt_price_now;

    // set event
    event PriceChanged(uint256 newPrice);
    event PriceChangedNow(uint256 newPrice, uint256 timestamp);

    // initialize function
     function initialize() public initializer {
        // set owner
        __Ownable_init(msg.sender);
        // set price
        btc_usdt_price[block.timestamp] = BTC_USDT_PRICE(0, block.timestamp);

        btc_usdt_price_now = BTC_USDT_PRICE(0, block.timestamp);
    }

    // set function to change btc_usdt price
    function changePriceNow(uint256 _price) public onlyOwner {
        // set price
        btc_usdt_price_now = BTC_USDT_PRICE(_price, block.timestamp);
        // emit event
        emit PriceChangedNow(_price, block.timestamp);
    }


    // get function to get btc_usdt price
    function getPriceNow() public view returns (uint256) {
        // return price
        return btc_usdt_price_now.price;
    }


}