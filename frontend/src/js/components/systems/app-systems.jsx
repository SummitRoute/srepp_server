var React = require('react');
var SystemsTable = require('./app-systems_table.jsx');

var Systems =
  React.createClass({
    render:function(){
      return (
          <div>
            <h2>Systems</h2>
            <SystemsTable/>
          </div>
        )
    }
  });
module.exports = Systems;
