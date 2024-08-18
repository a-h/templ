import React from "react";

export const Header = () => (<h1>React component Header</h1>);

export const Body = () => (<div>This is client-side content from React</div>);

export const Hello = (name: string) => (<div>Hello {name} (Client-side React, rendering server-side data)</div>);
