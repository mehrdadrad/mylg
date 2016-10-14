$(function() {
    var d = new Date(),
    host = "",
    stop = true,
    counter = 0,
    loopId,
    chart = c3.generate({
    bindto: '#chart',
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
function update() {
    var len = 0,
        host = $('#host').val();
        gap = [];
    if (host == "" || !stop) {
        return
    }
    if (host == undefined) {
        clearInterval(loopId);
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
    },"json");
}
loopId = setInterval(update, 1001);
});
