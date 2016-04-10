var React = require('react');
var Reqwest = require("reqwest");

var Help =
  React.createClass({
    getInitialState:function(){
			return {data:""}
		},
		componentWillMount:function(){
			Reqwest({
				url: '/api/help',
				type: '',
				success:function(resp){
					this.setState({data:resp})
				}.bind(this)
			})
		},
    render:function(){
      return (
          <div dangerouslySetInnerHTML={{__html: this.state.data}} />
        )
    }
  });
module.exports = Help;
