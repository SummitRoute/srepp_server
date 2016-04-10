var React = require('react');
var ProcessEventsTable = require('./app-process_events_table.jsx');

var ProcessEvents =
  React.createClass({
    render:function(){
      return (
            <div>
              <h2>Process Events</h2>
              <ProcessEventsTable/>
            </div>
        )
    }
  });
module.exports = ProcessEvents;
