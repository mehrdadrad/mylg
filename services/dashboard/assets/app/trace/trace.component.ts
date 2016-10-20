import { Component, HostListener, Injectable, Inject } from '@angular/core';
import { Http, Response, RequestOptions, URLSearchParams } from '@angular/http';
import { CommonModule } from '@angular/common';

import { GridOptions } from 'ag-grid/main';
import { Observable, Subscription } from 'rxjs/Rx'

@Component({
    selector: 'trace',
    templateUrl: './app/trace/trace.component.html',
})

@Injectable()
export class TraceComponent {
	public traceId
    public rowData
    public stats
    public lock
    public stop

    private gridOptions:GridOptions;

    subscription: Subscription;


    @HostListener('window:beforeunload', ['$event'])
    unloadHandler(event) {
        this.cleanUp()
    }

	ngOnInit(){
        this.traceErrorMsg = ""
		this.checked = 'checked'
        this.lock = false
        this.stop = false
	}

	ngAfterViewInit() {
		componentHandler.upgradeDom();
        this.gridOptions.api.sizeColumnsToFit()
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
        if (this.subscription) {
            this.subscription.unsubscribe()
            this.cleanUp()
        }
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
                width: 200
            },
            {
                headerName: "ASN",
                field: "ASN",
                width: 100,
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
                            this.rowData[idx].Best  = Math.min(this.rowData[idx].Elapsed, this.stats[idx].best)
                            this.rowData[idx].Worst = Math.max(this.rowData[idx].Elapsed, this.stats[idx].worst)
                            this.rowData[idx].Avg   = Math.round(((this.stats[idx].avg + this.rowData[idx].Elapsed) / 2) *1000) / 1000
                        }
                    }
                    this.rowData[idx].Loss = (this.stats[idx].loss*100)/this.stats[idx].sent

                    if (myTraceId == this.traceId) {
                        this.gridOptions.api.setRowData(this.rowData)
                        this.gridOptions.api.sizeColumnsToFit()
                        this.gridOptions.api.refreshView()
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
}
