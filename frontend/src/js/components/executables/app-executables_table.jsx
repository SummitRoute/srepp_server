var React = require('react');
var Reqwest = require("reqwest");
var Griddle = require('griddle-react');
var GriddleWithCallback = require('../GriddleWithCallback.jsx');
var Link = require('react-router-component').Link;
var Alert = require('react-bootstrap').Alert;

var Tokenizer = require('../react-typeahead/react-typeahead.js').Tokenizer;

var StructuredFilter = React.createClass({
  /**
   *
   */
  getDefaultProps: function(){
    return {
      // No properties
    }
  },


  /**
   *
   */
  handleChange: function(event){
      this.props.changeFilter(event.target.value);
  },


  /**
   *
   */
  render: function(){
    return <div className="filter-container input-group">
        <span className="input-group-addon">
          <i className="fa fa-search"></i>
        </span>
        <input type="text" name="filter" placeholder="Filter" className="form-control" onChange={this.handleChange} />
      </div>
  }
});


var ExecutablesTable = React.createClass({
  getInitialState: function() {
    return {
      errMsg: "",
      filter: "",
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
      url: '/api/files.json?start='+page*pageSize+"&length="+pageSize+"&filter="+filterString+"&sort="+sortColumn+"&sa="+sortAscending,
      type: 'json',
      success: function (resp) {
        var results = [];
        for (var i=0; i<resp.iTotalDisplayRecords; i++)
        {
          results[i] = {
            "Path" : Link({href:"/fileinfo?sha256="+resp.aaData[i]["Sha256"], children:resp.aaData[i]["FilePath"]}),
            "Last Seen": resp.aaData[i]["LastSeen"],
            "First Seen": resp.aaData[i]["FirstSeen"],
            "Product Name": resp.aaData[i]["ProductName"],
            "Company Name": resp.aaData[i]["CompanyName"],
            "Count": resp.aaData[i]["NumSystems"],
            "Signer": resp.aaData[i]["SignerSubjectShortName"]
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


  updateFilter: function(filter){
    // Set our filter to json data of the current filter tokens
    this.setState({filter: JSON.stringify(filter)});
  },


  // Some default values that get updated
  productOptions: function() {
    return ["Windows Search", "Windows Installer - Unicode", "VMware Tools", "Sysinternals Debugview", "Summit Route EPP", "Process Explorer",
    "Microsoft Windows Operating System", "Microsoft Visual Studio 2013", "Google Update", "Google Chrome", ""]
  },

  companyOptions: function() {
    return ["VMware, Inc.", "Sysinternals - www.sysinternals.com", "Sysinternals", "Summit Route", "Microsoft Corporation", "Google Inc.", ""]
  },

  signerOptions: function() {
    return ["Google Inc", "Microsoft Corporation", "Microsoft Windows", "VMware, Inc.", ""]
  },

  render: function(){
    if (this.state.errMsg != "") {
      return (
        <Alert bsStyle="danger">{this.state.errMsg}</Alert>
      );
    }

    return (
      <div>
        <Tokenizer
          placeholder=""
          options={[
            {category:"Path", type:"text"},
            {category:"LastSeen",type:"date"},
            {category:"FirstSeen", type:"date"},
            {category:"ProductName", type:"textoptions", options:this.productOptions},
            {category:"CompanyName", type:"textoptions", options:this.companyOptions},
            {category:"Count", type:"int"},
            {category:"Signer", type:"textoptions", options:this.signerOptions}]}
          customClasses={{
            input: "filter-tokenizer-text-input",
            results: "filter-tokenizer-list__container",
            listItem: "filter-tokenizer-list__item"
          }}
          updateFilter={this.updateFilter}
          onTokenAdd={function(filter) {this.updateFilter(filter);}}
          onTokenRemove={function(filter) {this.updateFilter(filter);}}
        />
        <GriddleWithCallback
          getExternalResults={this.getJsonData} filter={this.state.filter}
        />
      </div>
    )
  }
});
module.exports = ExecutablesTable;
