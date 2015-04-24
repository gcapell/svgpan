# svgpan
Simple javascript library to provide pan/zoom for viewing SVG in browser.

javascript from http://www.cyberz.org/blog/2009/12/08/svgpan-a-javascript-svg-panzoomdrag-library/
(i.e. https://code.google.com/p/svgpan/).  Copied to github just for safekeeping from
googlecode shutdown.

addsvgpan is a filter to add svgpan to an svg file.

It finds the first 'g' element, sets its 'id' attribute to 'viewport',
and inserts a <script> element before it.

sample usage: dot -Tsvg x.dot | addsvgpan > x.svg
