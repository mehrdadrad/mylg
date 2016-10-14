import { NgModule }      from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule }  from '@angular/router';

import { PingComponent } from './ping/ping';
import { HomeComponent } from './home/home';

import { AppComponent }  from './app.component';

@NgModule({
  imports: [
        RouterModule.forRoot([
			{ path: 'ping', component: PingComponent },
			{ path: 'home', component: HomeComponent },
			{ path: '', pathMatch: 'full', redirectTo:'/ping' },
		],{ useHash: true }),
		BrowserModule
  ],
  declarations: [ 
		AppComponent,
		PingComponent,
		HomeComponent
  ],
  bootstrap:    [ AppComponent ]
  })

export class AppModule { }
