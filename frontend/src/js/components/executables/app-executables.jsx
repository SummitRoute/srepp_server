var React = require('react');
var ExecutablesTable = require('./app-executables_table.jsx');

var Executables =
  React.createClass({
    render:function(){
      return (
            <div>
              <h2>Executables</h2>
              <ExecutablesTable/>
            </div>
        )
    }
  });
module.exports = Executables;
