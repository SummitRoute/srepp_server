var React = require('react');
var Reqwest = require("reqwest");

var PrivacyPolicy =
  React.createClass({
    getInitialState:function(){
			return {data:""}
		},
		componentWillMount:function(){
			Reqwest({
				url: '/api/privacy_policy',
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
module.exports = PrivacyPolicy;
