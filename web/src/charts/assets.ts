import * as d3 from 'd3'
import { assetColorBindings } from './constants'

const AssetChartFactory = (element): AssetChart => {
  let el = element

  // elements
  let svg
  let pie
  let arc
  let outerArc
  let slice
  let text
  let polyline

  // properties
  let width: number
  let height: number
  let radius: number
  let color

  // state
  let drawn = false

  const key: (d: any) => string = (d) => {
    return d.data.name
  }

  const midAngle: (d: any) => number = (d) => {
    return d.startAngle + (d.endAngle - d.startAngle)/2;
  }

  const draw: (h: number) => void = (h) => {
    svg = d3.select(el)
      .append("svg")
      .attr("width", "100%")
      .attr("height", "100%")
      .append("g")

    svg.append("g")
      .attr("class", "slices");
    svg.append("g")
      .attr("class", "labels");
    svg.append("g")
      .attr("class", "lines");

    const n = d3.select(el).node();
    width = n.getBoundingClientRect().width
    height = h
    radius = Math.min(width, height) / 2;

    pie = d3.pie()
      .sort(null)
      .value(function(d) {
        return d.nominal;
      });

    arc = d3.arc()
      .outerRadius(radius * 0.8)
      .innerRadius(radius * 0.4);

    outerArc = d3.arc()
      .innerRadius(radius * 0.9)
      .outerRadius(radius * 0.9);

    svg.attr("transform", "translate(" + width / 2 + "," + height / 2 + ")");

    color = d3.scaleOrdinal()
      .domain(Object.keys(assetColorBindings))
      .range(Object.keys(assetColorBindings).map(k => assetColorBindings[k]));

    drawn = true
  }

  const update: (items: AssetItem[]) => void = (items) => {

    /* ------- PIE SLICES -------*/
    slice = svg.select(".slices").selectAll("path.slice")
      .data(pie(items), key);

    slice.enter()
      .insert("path")
      .attr("fill", function(d) { return color(d.data.name); })
      .attr("class", "slice")
      .transition()
      .attr("d", function(d) { return arc(d) });

    slice.exit()
      .remove();

    /* ------- TEXT LABELS -------*/
    text = svg.select(".labels").selectAll("text")
      .data(pie(items), key);

    text.enter()
      .append("text")
      .style("fill", "#fff")
      .attr("dy", ".35em")
      .text(function(d) {
        return d.data.name + " " + d.data.amount;
      })
      .attr("transform", function(d) {
        const pos = outerArc.centroid(d);
        pos[0] = radius * (midAngle(d) < Math.PI ? 1 : -1);
        return "translate("+ pos +")";
      })
      .style("text-anchor", function(d) {
        return midAngle(d) < Math.PI ? "start":"end";
      });

    text.exit()
      .remove();

    /* ------- SLICE TO TEXT POLYLINES -------*/
    polyline = svg.select(".lines").selectAll("polyline")
      .data(pie(items), key);
    
    polyline.enter()
      .append("polyline")
      .attr("fill", "none")
      .attr("stroke", "#fff")
      .attr("points", function(d) {
        const pos = outerArc.centroid(d);
        pos[0] = radius * 0.95 * (midAngle(d) < Math.PI ? 1 : -1);
        return [arc.centroid(d), outerArc.centroid(d), pos];
      });
    
    polyline.exit()
      .remove();
  }

  const remove: () => void = () => {
    d3.select(el).select("svg").remove()
    drawn = false
  }

  return { draw, update, remove }
}

export default AssetChartFactory