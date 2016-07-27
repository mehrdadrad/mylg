$(function() {
    var d = new Date(),
    host = "",
    stop = true,
    counter = 0,
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
  zoom: {
    enabled: true
  },
  subchart: {
    show: true,
    length:0
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
    var len = 0,
        gap = [];
    if (host == "" || !stop) {
        return
    }
    if (counter == 0) {
        gap['x'] = ['x']
        gap['host'] = [host]    
        var t = new Date();
        for (var i=1;i<=10;i++) {
            t.setSeconds(t.getSeconds() - i);
            gap['host'].push(null);
            gap['x'].push(new Date (t));  
        }
        chart.flow({
                columns: [
                    gap['host'],
                    gap['x'],
                ],
        });

    }
    $.post( "/api/ping",{host: host},function( data ) {
        counter++;
        if (counter > 120) {
            len = 1;
        } else {
            len = 0;
        }
        var d = new Date()
            chart.flow({
                    columns: [
                        [host, data.rtt],
                        ['x', d],
                    ],
                    length:len,
            });
            if (data.rtt == 0) {
              chart.xgrids.add({value: d, text: 'timeout', class: 'red'});
            }
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
