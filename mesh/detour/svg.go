package detour

import (
	"bytes"
	"fmt"
	"lib_chaos/mesh"
)

func (t *Poly) points(nav *NavMesh) string {
	var (
		v0 = nav.MVert[t.Vs[0]]
		v1 = nav.MVert[t.Vs[1]]
		v2 = nav.MVert[t.Vs[2]]
	)
	return fmt.Sprintf("%f,%f %f,%f %f,%f", v0.X, v0.Z, v1.X, v1.Z, v2.X, v2.Z)
}

func (n *NavMesh) Svg(max mesh.Vert, s, t mesh.Vert, path []mesh.Vert) *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(head, max.Z, max.X))
	for i, t := range n.MPoly {
		var (
			v0 = n.MVert[t.Vs[0]]
			v1 = n.MVert[t.Vs[1]]
			v2 = n.MVert[t.Vs[2]]
			c  = mesh.VAdd(mesh.VAdd(v0, v1), v2)
			//color = []string{"blue", "orange", "green"}[t.GroupId%3]
		)
		//if nav.cache.tree[t.GroupId].big {
		//	color = "red"
		//}
		buf.WriteString(fmt.Sprintf("<polygon points=\"%s\" style=\"fill:none;stroke:blue;stroke-width:0.2\"/>\n", t.points(n)))
		buf.WriteString(fmt.Sprintf("<text x=\"%f\" y=\"%f\" font-size=\"1\" fill=\"black\">(%d)</text>", c.X/3, c.Z/3, i))
	}
	buf.WriteString(fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"0.4\" stroke=\"black\" stroke-width=\"0.05\" fill=\"red\" />\n", s.X, s.Z))
	buf.WriteString(fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"0.4\" stroke=\"black\" stroke-width=\"0.05\" fill=\"red\" />\n", t.X, t.Z))
	//buf.WriteString(fmt.Sprintf("<line x1=\"%f\" y1=\"%f\" x2=\"%f\" y2=\"%f\" style=\"stroke:red;stroke-width:0.25\" />\n",
	//	s.X, s.Z, t.X, t.Z))
	for i := 1; i < len(path); i++ {
		var v0, v1 = path[i-1], path[i]
		buf.WriteString(fmt.Sprintf("<line x1=\"%f\" y1=\"%f\" x2=\"%f\" y2=\"%f\" style=\"stroke:red;stroke-width:0.25\" />\n",
			v0.X, v0.Z, v1.X, v1.Z))
	}
	buf.WriteString("</g></svg>")
	return &buf
}

var head = "<svg  xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\" height=\"%f\" width=\"%f\">\n<script type=\"text/ecmascript\"><![CDATA[\n\"use strict\";\n\n/// CONFIGURATION\n/// ====>\n\nvar enablePan = 1; // 1 or 0: enable or disable panning (default enabled)\nvar enableZoom = 1; // 1 or 0: enable or disable zooming (default enabled)\nvar enableDrag = 0; // 1 or 0: enable or disable dragging (default disabled)\nvar zoomScale = 0.2; // Zoom sensitivity\n\n/// <====\n/// END OF CONFIGURATION\n\nvar root = document.documentElement;\n\nvar state = 'none', svgRoot = null, stateTarget, stateOrigin, stateTf;\n\nsetupHandlers(root);\n\n/**\n * Register handlers\n */\nfunction setupHandlers(root){\n\tsetAttributes(root, {\n\t\t\"onmouseup\" : \"handleMouseUp(evt)\",\n\t\t\"onmousedown\" : \"handleMouseDown(evt)\",\n\t\t\"onmousemove\" : \"handleMouseMove(evt)\",\n\t\t//\"onmouseout\" : \"handleMouseUp(evt)\", // Decomment this to stop the pan functionality when dragging out of the SVG element\n\t});\n\n\tif(navigator.userAgent.toLowerCase().indexOf('webkit') >= 0)\n\t\twindow.addEventListener('mousewheel', handleMouseWheel, false); // Chrome/Safari\n\telse\n\t\twindow.addEventListener('DOMMouseScroll', handleMouseWheel, false); // Others\n}\n\n/**\n * Retrieves the root element for SVG manipulation. The element is then cached into the svgRoot global variable.\n */\nfunction getRoot(root) {\n\tif(svgRoot == null) {\n\t\tvar r = root.getElementById(\"viewport\") ? root.getElementById(\"viewport\") : root.documentElement, t = r;\n\n\t\twhile(t != root) {\n\t\t\tif(t.getAttribute(\"viewBox\")) {\n\t\t\t\tsetCTM(r, t.getCTM());\n\n\t\t\t\tt.removeAttribute(\"viewBox\");\n\t\t\t}\n\n\t\t\tt = t.parentNode;\n\t\t}\n\n\t\tsvgRoot = r;\n\t}\n\n\treturn svgRoot;\n}\n\n/**\n * Instance an SVGPoint object with given event coordinates.\n */\nfunction getEventPoint(evt) {\n\tvar p = root.createSVGPoint();\n\n\tp.x = evt.clientX;\n\tp.y = evt.clientY;\n\n\treturn p;\n}\n\n/**\n * Sets the current transform matrix of an element.\n */\nfunction setCTM(element, matrix) {\n\tvar s = \"matrix(\" + matrix.a + \",\" + matrix.b + \",\" + matrix.c + \",\" + matrix.d + \",\" + matrix.e + \",\" + matrix.f + \")\";\n\n\telement.setAttribute(\"transform\", s);\n}\n\n/**\n * Dumps a matrix to a string (useful for debug).\n */\nfunction dumpMatrix(matrix) {\n\tvar s = \"[ \" + matrix.a + \", \" + matrix.c + \", \" + matrix.e + \"\\n  \" + matrix.b + \", \" + matrix.d + \", \" + matrix.f + \"\\n  0, 0, 1 ]\";\n\n\treturn s;\n}\n\n/**\n * Sets attributes of an element.\n */\nfunction setAttributes(element, attributes){\n\tfor (var i in attributes)\n\t\telement.setAttributeNS(null, i, attributes[i]);\n}\n\n/**\n * Handle mouse wheel event.\n */\nfunction handleMouseWheel(evt) {\n\tif(!enableZoom)\n\t\treturn;\n\n\tif(evt.preventDefault)\n\t\tevt.preventDefault();\n\n\tevt.returnValue = false;\n\n\tvar svgDoc = evt.target.ownerDocument;\n\n\tvar delta;\n\n\tif(evt.wheelDelta)\n\t\tdelta = evt.wheelDelta / 360; // Chrome/Safari\n\telse\n\t\tdelta = evt.detail / -9; // Mozilla\n\n\tvar z = Math.pow(1 + zoomScale, delta);\n\n\tvar g = getRoot(svgDoc);\n\n\tvar p = getEventPoint(evt);\n\n\tp = p.matrixTransform(g.getCTM().inverse());\n\n\t// Compute new scale matrix in current mouse position\n\tvar k = root.createSVGMatrix().translate(p.x, p.y).scale(z).translate(-p.x, -p.y);\n\n        setCTM(g, g.getCTM().multiply(k));\n\n\tif(typeof(stateTf) == \"undefined\")\n\t\tstateTf = g.getCTM().inverse();\n\n\tstateTf = stateTf.multiply(k.inverse());\n}\n\n/**\n * Handle mouse move event.\n */\nfunction handleMouseMove(evt) {\n\tif(evt.preventDefault)\n\t\tevt.preventDefault();\n\n\tevt.returnValue = false;\n\n\tvar svgDoc = evt.target.ownerDocument;\n\n\tvar g = getRoot(svgDoc);\n\n\tif(state == 'pan' && enablePan) {\n\t\t// Pan mode\n\t\tvar p = getEventPoint(evt).matrixTransform(stateTf);\n\n\t\tsetCTM(g, stateTf.inverse().translate(p.x - stateOrigin.x, p.y - stateOrigin.y));\n\t} else if(state == 'drag' && enableDrag) {\n\t\t// Drag mode\n\t\tvar p = getEventPoint(evt).matrixTransform(g.getCTM().inverse());\n\n\t\tsetCTM(stateTarget, root.createSVGMatrix().translate(p.x - stateOrigin.x, p.y - stateOrigin.y).multiply(g.getCTM().inverse()).multiply(stateTarget.getCTM()));\n\n\t\tstateOrigin = p;\n\t}\n}\n\n/**\n * Handle click event.\n */\nfunction handleMouseDown(evt) {\n\tif(evt.preventDefault)\n\t\tevt.preventDefault();\n\n\tevt.returnValue = false;\n\n\tvar svgDoc = evt.target.ownerDocument;\n\n\tvar g = getRoot(svgDoc);\n\n\tif(\n\t\tevt.target.tagName == \"svg\"\n\t\t|| !enableDrag // Pan anyway when drag is disabled and the user clicked on an element\n\t) {\n\t\t// Pan mode\n\t\tstate = 'pan';\n\n\t\tstateTf = g.getCTM().inverse();\n\n\t\tstateOrigin = getEventPoint(evt).matrixTransform(stateTf);\n\t} else {\n\t\t// Drag mode\n\t\tstate = 'drag';\n\n\t\tstateTarget = evt.target;\n\n\t\tstateTf = g.getCTM().inverse();\n\n\t\tstateOrigin = getEventPoint(evt).matrixTransform(stateTf);\n\t}\n}\n\n/**\n * Handle mouse button release event.\n */\nfunction handleMouseUp(evt) {\n\tif(evt.preventDefault)\n\t\tevt.preventDefault();\n\n\tevt.returnValue = false;\n\n\tvar svgDoc = evt.target.ownerDocument;\n\n\tif(state == 'pan' || state == 'drag') {\n\t\t// Quit pan mode\n\t\tstate = '';\n\t}\n}\n]]></script><g id=\"viewport\" transform=\"scale(0.5,0.5) translate(0,0)\">"
