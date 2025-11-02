<script setup lang="ts">
import type { Localization, Point } from "@/models/localization";
import { useStaticInfoStore } from "@/stores/static_info";
import { Map, View } from "ol";
import { applyStyle } from "ol-mapbox-style";
import { PMTilesVectorSource } from "ol-pmtiles";
import { FullScreen, MousePosition, ScaleLine, Zoom, ZoomToExtent } from "ol/control";
import { toStringHDMS } from "ol/coordinate";
import type { Extent } from "ol/extent";
import { GeoJSON } from "ol/format";
import { Vector as LayerVector, VectorTile } from "ol/layer";
import "ol/ol.css";
import { useGeographic } from "ol/proj";
import { Vector as SourceVector } from "ol/source";
import { Stroke, Style } from "ol/style";
import { storeToRefs } from "pinia";
import { onMounted } from "vue";

const props = defineProps<{
  localization: Localization;
}>();

const { staticInfo } = storeToRefs(useStaticInfoStore());

function makeGeoFeature(localization: Localization) {
  return {
    type: "FeatureCollection",
    features: [
      {
        type: "Feature",
        geometry: {
          type: "Polygon",
          coordinates: [
            [
              [
                localization.corner["upper-left"].coordinates.lon,
                localization.corner["upper-left"].coordinates.lat,
              ],
              [
                localization.corner["upper-right"].coordinates.lon,
                localization.corner["upper-right"].coordinates.lat,
              ],
              [
                localization.corner["lower-right"].coordinates.lon,
                localization.corner["lower-right"].coordinates.lat,
              ],
              [
                localization.corner["lower-left"].coordinates.lon,
                localization.corner["lower-left"].coordinates.lat,
              ],
            ],
          ],
        },
      },
    ],
  };
}

function makeBoundingBox(localization: Localization, marginRatio = 0.1): Extent {
  const pts = (Object.values(localization.corner) as Array<Point>).map((c) => c.coordinates);

  const lons = pts.map((p) => p.lon);
  const lats = pts.map((p) => p.lat);

  const minLon = Math.min(...lons);
  const maxLon = Math.max(...lons);
  const minLat = Math.min(...lats);
  const maxLat = Math.max(...lats);

  const width = maxLon - minLon;
  const height = maxLat - minLat;

  const dx = width * marginRatio;
  const dy = height * marginRatio;

  return [minLon - dx, minLat - dy, maxLon + dx, maxLat + dy];
}

onMounted(() => {
  useGeographic(); // To be able to center the map using lon/lat coordinates

  const carto = document.getElementById("carto");
  const cartoCoods = document.getElementById("carto-coords");

  if (!carto || !cartoCoods) {
    console.warn("All required elements for displaying map were not found.");
    return;
  }

  const vectorLayer = new VectorTile({
    declutter: true,
    source: new PMTilesVectorSource({
      url: staticInfo.value.pmtilesURL,
    }),
  });

  // Load styles for layer
  fetch(staticInfo.value.pmtilesStyleURL)
    .then((res) => res.json())
    .then((style) => {
      applyStyle(vectorLayer, style, "protomaps", { updateSource: false });
    });

  const cartoMap = new Map({
    target: "carto-map",
    layers: [vectorLayer],
    view: new View({
      center: [0, 0],
      zoom: 10,
    }),
    controls: [
      new FullScreen({
        source: carto,
      }),
      new MousePosition({
        coordinateFormat: (coords: number[] | undefined) => (coords ? toStringHDMS(coords, 2) : ""),
        projection: "EPSG:4326", // GCS WGS 84
        target: cartoCoods,
        className: "w-auto pointer-events-auto bg-white bg-opacity-75 p-1",
        placeholder: "-",
      }),
      new ScaleLine(),
      new Zoom(),
      new ZoomToExtent(),
    ],
  });

  const ROILayer = new LayerVector({
    source: new SourceVector({
      features: new GeoJSON().readFeatures(makeGeoFeature(props.localization)),
    }),
    style: new Style({
      stroke: new Stroke({
        color: "blue",
        width: 3,
        lineJoin: "bevel",
      }),
    }),
  });

  const ulLon = props.localization.corner["upper-left"].coordinates.lon;
  const urLon = props.localization.corner["upper-right"].coordinates.lon;
  const lrLon = props.localization.corner["lower-right"].coordinates.lon;
  const llLon = props.localization.corner["lower-left"].coordinates.lon;
  const ulLat = props.localization.corner["upper-left"].coordinates.lat;
  const urLat = props.localization.corner["upper-right"].coordinates.lat;
  const lrLat = props.localization.corner["lower-right"].coordinates.lat;
  const llLat = props.localization.corner["lower-left"].coordinates.lat;

  cartoMap.once("loadstart", () => {
    const centerLon = (ulLon + urLon + lrLon + llLon) / 4.0;
    const centerLat = (ulLat + urLat + lrLat + llLat) / 4.0;

    cartoMap.getView().setCenter([centerLon, centerLat]);

    cartoMap.getView().fit(makeBoundingBox(props.localization));
    cartoMap.updateSize();
  });

  cartoMap.once("loadstart", () => {
    cartoMap.addLayer(ROILayer);
  });
});
</script>

<template>
  <div id="carto" class="h-full">
    <div id="carto-map" class="h-full"></div>
    <div
      id="carto-coords"
      class="pointer-events-none flex -translate-y-[110%] justify-center"
    ></div>
    <div id="carto-overview" class="relative -translate-y-[110%]"></div>
  </div>
</template>

<style scoped></style>
