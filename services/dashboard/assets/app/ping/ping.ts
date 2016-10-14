import { Component } from '@angular/core';

var disabledPing = false,
	host = "";

@Component({
    selector: 'ping',
    templateUrl: './app/ping/ping.html'
})

export class PingComponent {
	ngOnInit(){
		jQuery.getScript('/js/ping.js');
		this.checked = 'checked'
	}
	ngAfterViewInit() {
		componentHandler.upgradeDom();
	}
	ngOnDestroy() {
		this.host = ""
	}
	onKey(event: any) {
	    host = event.target.value;
		this.host = host
	}
	onDisabledCheck() {
		if (disabledPing) {
			this.host = host
			disabledPing = false
		} else {
			this.host = ""
			disabledPing = true
		}
	}
}
