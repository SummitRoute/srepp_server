var React = require('react');
var Reqwest = require("reqwest");
var Griddle = require('griddle-react');
var Link = require('react-router-component').Link;

var ProcessEventsTable = React.createClass({
 	render: function(){
  var getJsonData = function(filterString, sortColumn, sortAscending, page, pageSize, callback) {
    Reqwest({
      url: '/api/processes.json?start='+page*pageSize+"&length="+pageSize+"&filter="+filterString+"&sort="+sortColumn+"&sa="+sortAscending,
      type: 'json',
      success: function (resp) {
        var results = [];
        for (var i=0; i<resp.iTotalDisplayRecords; i++)
        {
          results[i] = {
            "Path" : Link({href:"/fileinfo?sha256="+resp.aaData[i]["Sha256"], children:resp.aaData[i]["FilePath"]}),
            "Command" : resp.aaData[i]["CommandLine"],
            "Time": resp.aaData[i]["EventTime"]
          };
        }

        callback({
          results: results,
          totalResults: resp.iTotalRecords
        });
      }
    });
  };

    return (
      <Griddle getExternalResults={getJsonData} noDataMessage={"Loading data..."}
        tableClassName="table" resultsPerPage="25"
        showFilter={false} showSettings={true}/>
    )
  }
});
module.exports = ProcessEventsTable;
