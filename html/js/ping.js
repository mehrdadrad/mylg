$(function() {
    var d = new Date(),
    host = "",
    stop = true,
    chart = c3.generate({
    bindto: '#chart',
    size: {
        width: 800
    },
    data: {
        x: 'x',
        xFormat: '%M:%S',
        columns: [
        ],
      type: 'spline',
    },
point: {
        r: 1 
    },
    legend: {
        show: true
    },
    grid: {
        x: {
            show: true
        },
        y: {
            show: true
        }
    },    
    axis : {
        x : {
            label: {
                text: 'Time [days]'
            },
            type : 'timeseries',
            tick: {
                fit: true,
                count: 10,
                format: function (e, d) {
                   var format = d3.time.format("%H:%M:%S");
                   return format(e) 
                }
            }
      },
      y: {
          label: 'RTT [ms]',
          tick: {
              format: d3.format('.2f') 
          }    
      }
  },
  subchart: {
    show: true
  }
});
var gauge = c3.generate({
    bindto: '#gauge',
    data: {
         columns: [
             ['packet loss', 0]
         ],
         type: 'gauge' 
    },
    size: {
        height: 80
    }
})
function update() {
    if (host == "" || !stop) {
        return
    }
    console.log(host);
    $.post( "/api/ping",{host: host},function( data ) {
        var d = new Date()
            chart.flow({
                    columns: [
                        [host, data.rtt],
                        ['x', d],
                    ],
                    length:0,
                    grid: {
                        x: {
                            lines: [
                                {value: d, text: 'Lable 1'}
                            ]
                        }
                    },    
                });
            gauge.flow({
                columns:[['packet loss', data.pl]] 
            });
    },"json"); 
}

$('#pinghost').bind("enterKey",function(e){
    e.preventDefault();
    host = $('#pinghost').val();
});
$('#pinghost').keyup(function(e){
    e.preventDefault();
    if(e.keyCode == 13)
    {
          $(this).trigger("enterKey");
    }
});

$('#pingswitch').change(function () {
    var check = $(this).prop('checked');
    if (check) {
        stop = true
    } else {
        stop = false
    }
});
 
setInterval(update, 1001);

});
