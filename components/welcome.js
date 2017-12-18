import { Component } from 'preact';

export default class Welcome extends Component {
	render() {
		return (
			<section class="welcome">
				<h1>Paste your <strong>Local</strong> here...</h1>
				<p>Simply press <code>CTRL+V</code> (or <code>COMMAND+V</code> in macOS) and it'll fetch character's affiliations, threat level, kills, and losses.</p>
			</section>
		);
	}
}
