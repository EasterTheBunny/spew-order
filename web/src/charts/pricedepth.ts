import * as d3 from 'd3'

const PriceDepthChartFactory = (element: HTMLDivElement): PriceDepthChart => {
  let chartWidth = 700
  let chartHeight = 400

  // set the dimensions and margins of the graph
  let margin = {top: 0, right: 0, bottom: 0, left: 0}
  let el = element

  let svg
  let xScale
  let yScale
  let drawn = false
  let fillColor = "green"

  const draw: (width: number, height: number, color: string) => void = (width, height, color) => {
    fillColor = color
    chartWidth = width - margin.left - margin.right
    chartHeight = height - margin.top - margin.bottom
  
    xScale = d3.scaleLinear().range([chartWidth, 0]);
    yScale = d3.scaleBand().range([0, chartHeight]).padding(0.1);

    // append the svg object to the element
    svg = d3.select(el)
      .append("svg")
        .attr("width", width)
        .attr("height", height)
        .attr("background-color", "#030303")
      .append("g")
        .attr("transform",
              "translate(" + margin.left + "," + margin.top + ")")

    drawn = true
  }

  const update: (items: PriceDepthItem[], depth: number) => void = (items, depth) => {
    if (!drawn) {
      return
    }
    const prices: string[] = items.map(v => v.price)

    xScale.domain([0, depth]).nice() // depth
    yScale.domain(prices) // price

    svg.selectAll("g")
          .data(items, (d: PriceDepthItem) => d.price)
          .join(
            enter => {
              let g = enter.append("g")
                            .attr("transform", d => "translate(0,"+ yScale(d.price) +")")

              g.append("rect")
                    .attr("fill", fillColor)
                    .attr("width", (d: PriceDepthItem) => (chartWidth - xScale(d.depth)) / 2)
                    .attr("height", yScale.bandwidth())
                    .attr("opacity", "0.5")

              g.append("text")
                    .attr("class", "price")
                    .attr("x", d => 10)
                    .attr("y", d => yScale.bandwidth() / 2)
                    .attr("dy", ".35em")
                    .attr("font-size", "9px")
                    .text(d => d.price)

              g.append("text")
                    .attr("class", "depth")
                    .attr("x", d => 80)
                    .attr("y", d => yScale.bandwidth() / 2)
                    .attr("dy", ".35em")
                    .attr("font-size", "9px")
                    .text(d => d.depth)
            },
            update => update.call(update => {
              update.select("rect")
                      .transition()
                      .attr("width", (d: PriceDepthItem) => (chartWidth - xScale(d.depth)) / 2)
              
              update.select(".depth").text(d => d.depth)
            })
         )
  }

  const remove: () => void = () => {
    d3.select(el).select("svg").remove()
    drawn = false
  }

  return { draw, update, remove }
}

export default PriceDepthChartFactory
