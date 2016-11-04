import { Component } from '@angular/core';

declare var componentHandler
declare var jQuery:any;

var disabledPing = false,
	host = "";

@Component({
    selector: 'ping',
    templateUrl: './app/ping/ping.component.html'
})

export class PingComponent {
    public hideSubmitBtn
    public host
    public checked

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

    onFocus() {
        this.hideSubmitBtn = true
    }

    onBlur() {
        setTimeout(() => {
            this.hideSubmitBtn = false
        }, 600)
    }

	onKey(event: any, dest: string) {
        host = dest || event.target.value;
        this.host = host
        this.hideSubmitBtn = false
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
