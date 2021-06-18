import * as d3 from 'd3'
import dayjs from 'dayjs'

const PriceHistoryChartFactory = (element): PriceHistoryChart => {
  let chartWidth = 700
  let chartHeight = 400

  let margin = {top: 5, right: 10, bottom: 20, left: 50}
  let el = element
  
  let svg
  let xScale
  let yScale
  let xAxis
  let yAxis
  let clip
  let brush
  let line
  let idleTimeout
  let data: CandleItem[]
  let xLabelTransform = "translate(10,0)"
  let focus
  let focusText
  let drawn = false
  
  const idled = () => { idleTimeout = null; }
  
  const formatMajor = (d: Date, i) => {
    if (i%3 === 0) {
      return dayjs(d).format('M/D')
    } else {
      return ''
    }
  }
  
  const buildAxis = () => {
    return d3.axisBottom(xScale)
              .tickSize(10)
              .tickFormat(formatMajor)
  }

  const reloadBrush = (event) => {
    // What are the selected boundaries?
    let extent = event.selection

    // If no selection, back to initial coordinate. Otherwise, update X axis domain
    if(!extent){
      // This allows to wait a little bit
      if (!idleTimeout) return idleTimeout = setTimeout(idled, 350);
      xScale.domain([ 4,8])
    }else{
      xScale.domain([ xScale.invert(extent[0]), xScale.invert(extent[1]) ])
      // This remove the grey brush area as soon as the selection has been done
      line.select(".brush").call(brush.move, null)
    }

    // Update axis and line position
    xAxis.transition()
          .duration(500)
          .call(buildAxis())
            .selectAll("text")
            .attr("transform", xLabelTransform)
            .style("text-anchor", "end")
    
    line.select('.line')
        .transition()
        .duration(700)
        .attr("d", lineGenerator(xScale, yScale))
        
    xAxis.selectAll("g")
        .filter(function (d, i) {
          return i%6!=0;
        })
        .classed("minor", true)
  }
  
  const bisectDate = d3.bisector(function(d: any, x: any) { return d.time - x; }).left;
  
  const mouseover = (d) => {
    focus.style("opacity", 1)
    focusText.style("opacity",1)
  }
  
  const mousemove = (event) => {
    event.preventDefault()
    const pointer = d3.pointer(event, this)
    const xm = xScale.invert(pointer[0])
    const ym = yScale.invert(pointer[1]-margin.bottom)
    const i = bisectDate(data, xm)
    // recover coordinate we need
    
    let selectedData = data[i]
    focus
      .attr("cx", xScale(selectedData.time))
      .attr("cy", yScale(selectedData.close))
    focusText
      .html(d3.format(",.2r")(selectedData.close))
      .attr("x", xScale(selectedData.time)+15)
      .attr("y", yScale(selectedData.close))
      
  }
  
  const mouseleave = (d) => {
    focus.style("opacity", 0)
    focusText.style("opacity", 0)
  }
  
  const draw: (width: number, height: number) => void = (width, height) => {
    const n = d3.select(el).node();
    width = n.getBoundingClientRect().width

    chartWidth = width - margin.left - margin.right
    chartHeight = height - margin.top - margin.bottom
  
    xScale = d3.scaleTime().range([0, chartWidth])
    yScale = d3.scaleLinear().range([chartHeight, 0])

    // append the svg object to the body of the page
    svg = d3.select(el)
      .append("svg")
        .attr("width", width)
        .attr("height", height)
      .append("g")
        .attr("transform",
              "translate(" + margin.left + "," + margin.top + ")")

    xAxis = svg.append("g")
                .attr("id", "x_axis")
                .attr("class", "x axis")
                .attr("transform", "translate(0," + chartHeight + ")")
    
    yAxis = svg.append("g")
    
    // Add a clipPath: everything out of this area won't be drawn.
    clip = svg.append("defs")
              .append("svg:clipPath")
              .attr("id", "clip")
              .append("svg:rect")
              .attr("width", chartWidth )
              .attr("height", chartHeight )
              .attr("x", 0)
              .attr("y", 0)
    
              // Add brushing
    brush = d3.brushX()                   // Add the brush feature using the d3.brush function
              .extent( [ [0,0], [chartWidth, chartHeight] ] )  // initialise the brush area: start at 0,0 and finishes at width,height: it means I select the whole graph area
              .on("end", reloadBrush)               // Each time the brush selection changes, trigger the 'updateChart' function

    line = svg.append('g').attr("clip-path", "url(#clip)")
    
    focus = svg
      .append('g')
      .append('circle')
      .style("fill", "steelblue")
      .attr("stroke", "#fff")
      .attr("stroke-width", 2)
      .attr('r', 4.5)
      .style("opacity", 0)
    
    focusText = svg
      .append('g')
      .append('text')
      .style("opacity", 0)
      .attr("text-anchor", "left")
      .attr("alignment-baseline", "middle")

    // If user double click, reinitialize the chart
    svg.on("dblclick",function(){
      xScale.domain(d3.extent(data, d => d.time))
      xAxis.transition().call(buildAxis())
            .selectAll("text")
            .attr("transform", xLabelTransform)
            .style("text-anchor", "end")
      line.select('.line')
          .transition()
          .attr("d", lineGenerator(xScale, yScale))
          
      xAxis.selectAll("g")
        .filter(function (d, i) {
          return i%6!=0;
        })
        .classed("minor", true)
    })
    
    svg.append('rect')
        .style("fill", "none")
        .style("pointer-events", "all")
        .attr('width', chartWidth)
        .attr('height', chartHeight)
        .on('mouseover', mouseover)
        .on('mousemove', mousemove)
        .on('mouseout', mouseleave)

    drawn = true
  }

  const update: (items: CandleItem[]) => void = (items) => {
    if (!drawn) {
      return
    }
    data = items

    xScale.domain(d3.extent(items, d => d.time))
    yScale.domain(d3.extent(items, d => d.close)).nice()

    xAxis.call(buildAxis())
          .selectAll("text")
          .attr("transform", xLabelTransform)
          .style("text-anchor", "end")
    yAxis.call(d3.axisLeft(yScale).ticks(5))

    // Add the line
    line.append("path")
        .datum(items)
        .attr("class", "line")  // I add the class line to be able to modify this line later on.
        .attr("fill", "none")
        .attr("stroke", "steelblue")
        .attr("stroke-width", 1.5)
        .attr("d", lineGenerator(xScale, yScale))

    line.append("g")
        .attr("class", "brush")
        .call(brush)
    
    xAxis.selectAll("g")
        .filter(function (d, i) {
          return i%6!=0;
        })
        .classed("minor", true)
  }

  const lineGenerator = (x, y) => {
    return d3.line<CandleItem>()
              .curve(d3.curveCardinal)
              .x(d => x(d.time))
              .y(d => y(d.close))
  }

  const remove: () => void = () => {
    d3.select(el).select("svg").remove()
    drawn = false
  }

  return { draw, update, remove }
}


export default PriceHistoryChartFactory