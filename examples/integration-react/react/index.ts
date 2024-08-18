import { createRoot } from 'react-dom/client';
import { Header, Body, Hello } from './components';

// Render the React component into the templ page at the react-header.
const headerRoot = document.getElementById('react-header');
if (!headerRoot) {
	throw new Error('Could not find element with id react-header');
}
const headerReactRoot = createRoot(headerRoot);
headerReactRoot.render(Header());

// Add the body React component.
const contentRoot = document.getElementById('react-content');
if (!contentRoot) {
	throw new Error('Could not find element with id react-content');
}
const contentReactRoot = createRoot(contentRoot);
contentReactRoot.render(Body());

// Provide a helper for rendering hello.
export function renderHello(e: HTMLElement) {
	const name = e.getAttribute('data-name') ?? "";
	createRoot(e).render(Hello(name));
}
