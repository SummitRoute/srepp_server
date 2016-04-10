var React = require('react');
var Reqwest = require("reqwest");
var Griddle = require('griddle-react');
var GriddleWithCallback = require('../GriddleWithCallback.jsx');
var Link = require('react-router-component').Link;
var Alert = require('react-bootstrap').Alert;

var SystemsTable = React.createClass({
  getInitialState: function() {
    return {
      errMsg: "",
    }
  },

  getJsonData: function(filterString, sortColumn, sortAscending, page, pageSize, callback) {
    thisComponent = this;

    if (filterString==undefined) {
      filterString = "";
    }
    if (sortColumn==undefined) {
      sortColumn = "";
    }

    Reqwest({
      url: '/api/systems.json?start='+page*pageSize+"&length="+pageSize+"&filter="+filterString+"&sort="+sortColumn+"&sa="+sortAscending,
      type: 'json',
      success: function (resp) {
        var results = [];
        for (var i=0; i<resp.iTotalDisplayRecords; i++)
        {
          results[i] = {
            "Machine Name" : Link({href:"/systeminfo?uuid="+resp.aaData[i]["System"], children:resp.aaData[i]["MachineName"]}),
            "Comment": resp.aaData[i]["Comment"],
            "Agent Version": resp.aaData[i]["AgentVersion"],
            "OS": resp.aaData[i]["OSHumanName"],
            "Manufacturer": resp.aaData[i]["Manufacturer"],
            "Model": resp.aaData[i]["Model"],
            "Last Seen": resp.aaData[i]["LastSeen"],
            "First Seen": resp.aaData[i]["FirstSeen"]
          };
        }

        callback({
          results: results,
          totalResults: resp.iTotalRecords,
          pageSize: pageSize
        });
      },
      error: function (err) {
        var msg = "The server was unable to handle that request.";
        thisComponent.setState({errMsg:msg});
      }
    });
  },

 	render: function(){
    if (this.state.errMsg != "") {
      return (
       <Alert bsStyle="danger">{this.state.errMsg}</Alert>
      );
    }

    return (
      <GriddleWithCallback
        getExternalResults={this.getJsonData}
        showFilter={false}
      />
    )
  }
});
module.exports = SystemsTable;
