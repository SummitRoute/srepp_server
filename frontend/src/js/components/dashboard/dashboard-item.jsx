var React = require('react');
var MicroBarChart = require('react-micro-bar-chart');
var _string = require('underscore.string');

var Dashboard =
  React.createClass({
    getDefaultProps: function(){
      return {
        title: '0 new events',
        data: [0,0,0,0,0,0,0],
        tooltipText: "%d events"
      }
    },

    tooltipTemplate: function (d, i, data){
      return _string.sprintf(this.props.tooltipText, d)
    },

    render:function(){
      return (
        <div className="row">
          <div className="col-md-6">
            <p>{this.props.title}</p>
          </div>
          <div className="col-md-6 barchart">
            <MicroBarChart
              width="150" height="40"
              data={this.props.data}
              tooltip="true"
              tipOffset={[0,20]}
              tipTemplate={this.tooltipTemplate}
              hoverColor="#80B5FF"
              fillColor="#2b82ff"
            />
          </div>
        </div>
        )
    }
  });
module.exports = Dashboard;
