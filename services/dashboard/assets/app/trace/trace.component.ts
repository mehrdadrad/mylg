import { Component, HostListener, Injectable, Inject } from '@angular/core';
import { Http, Response, RequestOptions, URLSearchParams } from '@angular/http';
import { CommonModule } from '@angular/common';

import { GridOptions } from 'ag-grid/main';
import { Observable, Subscription } from 'rxjs/Rx'

declare var c3: any;

@Component({
    selector: 'trace',
    templateUrl: './app/trace/trace.component.html',
})

@Injectable()
export class TraceComponent {
	public traceId
    public rowData
    public chartData
    public stats
    public lock
    public stop
    public proto
    public chart
    public jChart
    public jLast
    public geoUpdated

    private gridOptions:GridOptions;

    subscription: Subscription;

    @HostListener('window:beforeunload', ['$event'])
    unloadHandler(event) {
        this.cleanUp()
    }

	ngOnInit(){
        jQuery.getScript('/js/trace.js');
        this.traceErrorMsg = ""
		this.checked = 'checked'
        this.lock = false
        this.stop = false
        this.proto = "icmp"
		this.geoUpdated = false
		this.chartData = []
	}

	ngAfterViewInit() {
		componentHandler.upgradeDom();
        this.gridOptions.api.sizeColumnsToFit()
        this.initRTTChart()
        this.initJitterChart()
	}

    ngOnDestroy() {
        this.cleanUp()
    }

    onFocus() {
        this.hideSubmitBtn = true
    }

    onBlur() {
        setTimeout(() => {
            this.hideSubmitBtn = false
        }, 600)
    }


	onKey(event: any, dest: string) {
	    let args = dest || event.target.value;
        args = args + this.options()
        if (this.subscription) {
            this.subscription.unsubscribe()
            this.cleanUp()
        }
        this.initRTTChart()
        this.initJitterChart()
        if (event.target.value) {
            event.target.blur()
        }
        this.hideSubmitBtn = false
        this.gridOptions.rowData = []
        this.stats = new Array(64)
        this.getTraceID(args).subscribe(
			data => {
				this.traceId = data.id
                if (data.err != "") {
                    this.traceErrorMsg = data.err
                } else {
                    this.traceErrorMsg = ""
                }
			},
			err => console.error(err),
			() => {
                if (this.traceId == -1) {
                    return
                }
                 this.subscription = Observable.interval(100)
            .       subscribe(number => {
                        if (!this.lock && !this.stop) {
                            this.getTraceData()
                        }
                    })
			}
		);
    }

    onClickCheck() {
        if (this.stop) {
            this.stop = false
        } else {
            this.stop = true
        }
    }

    setProto(p) {
        this.proto = p
    }

    options(){
        let ops = [" "]
        switch(this.proto) {
            case "udp":
                ops.push("-u")
                break
            case "tcp":
                ops.push("-t")
                break
        }
        return ops.join(" ")
    }

	constructor(@Inject(Http) private http : Http) {
        this.gridOptions = <GridOptions>{};
		this.gridOptions.rowData = this.createEmptyRows();
		this.gridOptions.columnDefs = this.createColumnDefs();
    }

    private createColumnDefs() {
        return [
            {headerName: "Hop", field: "row", width: 40},
            {
                headerName: "IP/Host",
                field: "IP",
                width: 300,
            },
            {
                headerName: "Holder",
                field: "Holder",
                width: 140,
                cellRenderer: function (params) {
                    if (!params.value) {
                        return ""
                    }
                    var r = params.value.split(" - ")
                    if (r.length > 0) {
                        return r[0]
                    } else {
                        return params.value
                    }
                }
            },
            {
                headerName: "ASN",
                field: "ASN",
                width: 70,
                cellRenderer: function (params) {
                    if (params.value && params.value != 0) {
                        return params.value
                    } else {
                        return ''
                    }
                }
            },
            {
                headerName: "Loss",
                field: "Loss",
                width: 55,
                cellRenderer: function (params) {
                    if (params.value == '-') {
                        return ''
                    } else if (params.value) {
                        return Math.round(params.value*10)/10 + ' %'
                    } else {
                        return '0 %'
                    }
                }
            },
            {
                headerName: "Sent",
                field: "Sent",
                width: 55,
                cellClass: 'cell-number',
            },
            {
                headerName: "Last",
                field: "Elapsed",
                width: 80,
                cellClass: 'cell-number',
                cellRenderer: function (params) {
                    if (params.value && params.value != 0) {
                        return params.value
                    } else {
                        return ''
                    }
                }
            },
            {
                headerName: "Avg",
                field: "Avg",
                width: 80,
                cellClass: 'cell-number',
                cellRenderer: function (params) {
                    if (params.value && params.value != 0) {
                        return params.value
                    } else {
                        return ''
                    }
                }
            },
            {
                headerName: "Best",
                field: "Best",
                width: 80,
                cellClass: 'cell-number',
                cellRenderer: function (params) {
                    if (params.value && params.value != 0) {
                        return params.value
                    } else {
                        return ''
                    }
                }
            },
            {
                headerName: "Worst",
                field: "Worst",
                width: 80,
                cellClass: 'cell-number',
                cellRenderer: function (params) {
                    if (params.value && params.value != 0) {
                        return params.value
                    } else {
                        return ''
                    }
                }
            }
        ];
    }

    private createEmptyRows() {
        let rowData:any[] = [];
        for (var i = 0; i < 15; i++) {
            rowData.push({
                Loss: '-'
            },
		    );
        }
        return rowData;
    }

	closeTrace(): Observable {
        let params = "id=" + this.traceId
        let options = new RequestOptions({
                search: new URLSearchParams('id='+this.traceId)
        });
        return this.http.get('api/close.trace', options)
			.map((res:Response) => res.json())
    }

	getTraceID(args): Observable{
		let options = new RequestOptions({
            search: new URLSearchParams('a='+args)
        });
		return this.http.get('api/init.trace', options)
			.map((res:Response) => res.json())
				}

	getTraceData(): Observable {
		let params = "id=" + this.traceId
        let myTraceId = this.traceId
		let options = new RequestOptions({
            search: new URLSearchParams('id='+this.traceId)
        });
        if (!this.traceId) {
            console.log("trace id is not available")
            return
        }
        this.lock = true
		return this.http.get('api/get.trace', options)
			.map((res:Response) => res.json())
			.subscribe(
				data => {
                    if (myTraceId != this.traceId) {
                        return
                    }
                    let idx = data.Id -1
                    this.rowData = this.gridOptions.rowData
                    if (data.Hop.length > 2) {
                        data.IP = data.Hop
                    }
                    data.row = data.Id
                    this.rowData[idx] = data

                    if (this.stats[idx] == undefined) {
                        this.stats[idx] = { sent: 0, loss: 0, avg: 0, best:0, worst:0 }
                    }

                    this.stats[idx].sent++
                    this.rowData[idx].Sent = this.stats[idx].sent
                    if (this.rowData[idx].Elapsed == 0) {
                        this.stats[idx].loss++
                    } else {
                        if (this.stats[idx].sent == 1) {
                            this.stats[idx].avg   = this.rowData[idx].Elapsed
                            this.stats[idx].best  = this.rowData[idx].Elapsed
                            this.stats[idx].worst = this.rowData[idx].Elapsed

                            this.rowData[idx].Avg   = this.stats[idx].avg
                            this.rowData[idx].Best  = this.stats[idx].best
                            this.rowData[idx].Worst = this.stats[idx].worst
                        } else {
                            this.stats[idx].best  = Math.min(this.rowData[idx].Elapsed, this.stats[idx].best)
                            this.stats[idx].worst = Math.max(this.rowData[idx].Elapsed, this.stats[idx].worst)
                            this.stats[idx].avg   = Math.round(((this.stats[idx].avg + this.rowData[idx].Elapsed) / 2) *1000) / 1000

                            this.rowData[idx].Avg   = this.stats[idx].avg
                            this.rowData[idx].Best  = this.stats[idx].best
                            this.rowData[idx].Worst = this.stats[idx].worst
                        }
                    }
                    this.rowData[idx].Loss = (this.stats[idx].loss*100)/this.stats[idx].sent

                    if (myTraceId == this.traceId) {
                        this.gridOptions.api.setRowData(this.rowData)
                        this.gridOptions.api.sizeColumnsToFit()
                        this.gridOptions.api.refreshView()
                    }

					if (data.IP.length > 5) {
						this.chartData.push([data.IP, this.rowData[idx].Elapsed])
					}

					if (data.Last) {
						this.updateRTTChart(this.chartData)
						this.updateJitterChart(this.rowData[idx].Elapsed)
						this.chartData = []
						if (!this.geoUpdated) {
							this.updateGeo(data.IP)
						}
					}
				},
				err => {
                    this.lock = false
                    console.error(err)
                },
				() => {
                    this.lock = false
				}
			);
	}

    cleanUp() {
		this.geoUpdated = false
        if (this.subscription) {
            this.subscription.unsubscribe()
        }
        if (this.traceId) {
            this.closeTrace().subscribe(() => {
                this.traceId = undefined
            })
        } else {
           this.traceId = undefined
        }
    }

	updateGeo(host) {
		let options = new RequestOptions({
            search: new URLSearchParams('ip='+host)
        });
		this.http.get('api/geo', options)
			.map((res:Response) => res.json())
			.subscribe(
				data => {
					this.geoFrom = data.CitySrc + ' '+ data.CountrySrc
					this.geoTo = data.CityDst + ' '+ data.CountryDst
					this.geoUpdated = true
				},
			)
	}

    initRTTChart() {
        this.chart = c3.generate({
			bindto: '#rChart',
			data: {
                x: 'x',
                xFormat: '%M:%S',
                columns: [],
                type: 'spline',
            },
			axis : {
					x : {
						extent: function() {
							var t = new Date();
							t.setSeconds(t.getSeconds() - 20);
							return [new Date(t), new Date()];
						},
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
			},
			legend: {
				show: true,
			},
			grid: {
				x: {
					show: true
				},
				y: {
					show: true
				}
			},
        })
    }

    updateRTTChart(data) {
		if (this.chart.data().length == 0) {
			var gap = []
			gap['x'] = ['x']
			gap['host'] = [data[0][0]]
			var t = new Date();
			for (var i=1;i<=25;i++) {
				t.setSeconds(t.getSeconds() - i);
				gap['host'].push(null);
				gap['x'].push(new Date (t));
			}
			this.chart.flow({
					columns: [
						gap['host'],
						gap['x'],
					],
			});
		}
		var date = new Date()
		var cols = [['x', date]]
			for (var d of data) {
				cols.push(d)
			}
            this.chart.flow({
                    columns: cols,
                    length:0,
            });
	}

    initJitterChart() {
        this.jChart = c3.generate({
			bindto: '#jChart',
			data: {
                x: 'x',
                xFormat: '%M:%S',
                columns: [],
                type: 'spline',
            },
            size: {
                height: 120
            },
            axis: {
            	x : {
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
                    label: 'Jitter [ms]',
                    tick: {
                        count: 4,
                        format: d3.format('.2f')
                    },
                    min: 0
                }
            },
            point: {
                show: false
            },
            legend: {
                show: false
            }
        })
    }

    updateJitterChart(data) {
		if (this.jChart.data().length == 0) {
			var gap = []
			gap['x'] = ['x']
			gap['host'] = ['jitter']
			var t = new Date();
			for (var i=1;i<=25+1;i++) {
				t.setSeconds(t.getSeconds() - i);
				gap['host'].push(null);
				gap['x'].push(new Date (t));
			}
			this.jChart.flow({
					columns: [
						gap['host'],
						gap['x'],
					],
			});
		}
        if (!this.jLast) {
            this.jLast = data
            return
        }
		var date = new Date()
        this.jChart.flow({
            columns: [
                ['jitter', Math.abs(this.jLast - data)],
                ['x', date],
            ],
            length:0,
        });
        this.jLast = data
	}
}
