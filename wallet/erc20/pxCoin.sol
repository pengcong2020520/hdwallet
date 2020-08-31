pragma solidity ^0.5.1;

import './erc20.sol';

contract pxCoin is ERC20 {
    
    //coin name
    string public name = "pxcb";  //coin name 
    string public symbol = "pxc"; // coin tag 
    
    //defined owner
    address public fundation;
    address public issuer;
    
    //defined sum amount
    uint private _totalSupply;
    //uint public _totalAirDrop;

    //defined balance
    mapping(address=>uint) _balance;
    mapping(address=>mapping(address=>uint)) _allowance;
    
    event Transfer(address indexed _from, address indexed _to, uint _value);
    event Approval(address indexed _owner, address indexed _spender, uint _value);
    
    //constructor init data
    constructor (uint totalSupply, address payable _owner) public payable{
        _totalSupply = totalSupply;
        fundation = _owner;
        _balance[fundation] = totalSupply * 20 / 100;
        issuer = msg.sender;
        _balance[issuer] = totalSupply * 80 / 100;
    }
    
    //defined AirDrop
    // function airDrop(address _to, uint _value) public returns (bool){
    //     assert(msg.sender == fundation);
    //     if (_totalAirDrop + _value + _balance[fundation] > 0 &&
    //         _totalAirDrop + _value + _balance[fundation] < _totalSupply &&
    //         address(0) != _to
    //         ){
            
    //         _balance[_to] += _value;
    //         return true;
    //     }
    //     else {
    //         return false;
    //     }
    // }
    
    
    function totalSupply() public view returns (uint totalSupply) {
        totalSupply = _totalSupply;
        return totalSupply;
    }
    
    function balanceOf(address _owner) public view returns (uint balance) {
        return _balance[_owner];
    }
    
    //address(msg.sender) transfer to address(_to) 
    function transfer(address _to, uint _value) public returns (bool success) {
        if (_balance[msg.sender] >= _value && 
            _balance[_to] + _value > 0 &&
            _to != address(0)
        ){
            
            _balance[msg.sender] -= _value;
            _balance[_to] += _value;
            
            emit Transfer(msg.sender, _to, _value);
            
            return true;
            
        }
        else {
            
            return false;
        }
    }
    
    function transferFrom(address _from, address _to, uint _value) public returns (bool success) {
        
        if (_balance[_from] >= _value &&
            _balance[_to] + _value > 0 && 
            _allowance[_from][_to] >= _value && 
            address(0) != _to
        ){
            
            _allowance[_from][_to] -= _value;
            _balance[_to] += _value;
            _balance[_from] -= _value;
            
            return true;
        }
        else {
            return false;
        }
        
    }
    
    function approve(address _spender, uint _value) public returns (bool success) {
        if ( _balance[msg.sender] >= _value && 
             _balance[_spender] + _value > 0 && 
             address(0) != _spender
        ){
            
            _allowance[msg.sender][_spender] = _value;
            
            emit Approval(msg.sender, _spender, _value);
            return true;
        }
        else {
            return false;
        }
    }
    
    function allowance(address _owner, address _spender) public view returns (uint remaining) {
        return _allowance[_owner][_spender];
    }
  
    function getAddr() view public returns(address) {
        return address(this);
    }
    
}
