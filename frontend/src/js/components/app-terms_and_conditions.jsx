var React = require('react');
var Reqwest = require("reqwest");

var TermsAndConditions =
  React.createClass({
    getInitialState:function(){
			return {data:""}
		},
		componentWillMount:function(){
			Reqwest({
				url: '/api/terms_and_conditions',
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
module.exports = TermsAndConditions;
