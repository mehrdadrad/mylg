import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { HttpModule, JsonpModule } from '@angular/http';

import { PingComponent } from './ping/ping.component';
import { TraceComponent } from './trace/trace.component';
import { HomeComponent } from './home/home.component';
import { AgGridModule } from '../js/vendor/ag-grid-ng2/main';

import { AppComponent } from './app.component';

@NgModule({
  imports: [
        RouterModule.forRoot([
			{ path: 'ping',  component: PingComponent  },
			{ path: 'trace', component: TraceComponent },
			{ path: 'home',  component: HomeComponent  },
			{ path: '', pathMatch: 'full', redirectTo:'/trace' },
		],{ useHash: true }),
		BrowserModule,
        AgGridModule.withNg2ComponentSupport(),
        HttpModule,
        JsonpModule,
  ],
  declarations: [
		AppComponent,
		PingComponent,
		TraceComponent,
		HomeComponent
  ],
  bootstrap: [ AppComponent ]
  })

export class AppModule { }
