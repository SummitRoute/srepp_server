// Custom JS for SREPP

window.setTimeout(function() {
  $(".flash").fadeTo(500, 0).slideUp(500, function(){
      $(this).remove();
  });
}, 3000);


function getUrlParameter(sParam)
{
    var sPageURL = window.location.search.substring(1);
    var sURLVariables = sPageURL.split('&');
    for (var i = 0; i < sURLVariables.length; i++)
    {
        var sParameterName = sURLVariables[i].split('=');
        if (sParameterName[0] == sParam)
        {
            return sParameterName[1];
        }
    }
}
