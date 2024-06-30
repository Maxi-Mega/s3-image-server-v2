<script setup lang="ts">
import { onMounted } from "vue";
import { View } from "ol";
import Map from "ol/Map";
import TileLayer from "ol/layer/Tile";
import { XYZ } from "ol/source";
import { FullScreen, MousePosition, ScaleLine, Zoom, ZoomToExtent } from "ol/control";
import { toStringHDMS } from "ol/coordinate";
import { useGeographic } from "ol/proj";
import { Vector as LayerVector } from "ol/layer";
import { Vector as SourceVector } from "ol/source";
import { GeoJSON } from "ol/format";
import { Stroke, Style } from "ol/style";
import type { Localization } from "@/models/localization";
import { useStaticInfoStore } from "@/stores/static_info";
import "ol/ol.css";

const props = defineProps<{
  localization: Localization;
}>();

const staticInfo = useStaticInfoStore();

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

onMounted(() => {
  useGeographic(); // To be able to center the map using lon/lat coordinates

  const carto = document.getElementById("carto");
  const cartoCoods = document.getElementById("carto-coords");

  if (!carto || !cartoCoods) {
    console.warn("All required elements for displaying map were not found.");
    return;
  }

  const cartoMap = new Map({
    target: "carto-map",
    layers: [
      new TileLayer({
        source: new XYZ({
          url: staticInfo.staticInfo.tileServerURL,
        }),
      }),
    ],
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

    const marginPercent = 10;
    const deltaX = (Math.abs(ulLon - urLon) * marginPercent) / 100;
    const deltaY = (Math.abs(ulLat - llLat) * marginPercent) / 100;
    const cartoExtent = [
      Math.min(ulLon, llLon) - deltaX,
      Math.min(llLat, lrLat) - deltaY,
      Math.max(urLon, lrLon) + deltaX,
      Math.max(ulLat, urLat) + deltaY,
    ];

    cartoMap.getView().fit(cartoExtent);
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
