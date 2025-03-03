import * as datastar from "@starfederation/datastar/dist/";
// Refer following links for officeial plugins
// https://data-star.dev/bundler
// https://github.com/starfederation/datastar/blob/main/library/src/bundles/datastar.ts
import { GET } from "@starfederation/datastar/dist/plugins/official/backend/actions/get";
import { DELETE } from "@starfederation/datastar/dist/plugins/official/backend/actions/delete";
import { PATCH } from "@starfederation/datastar/dist/plugins/official/backend/actions/patch";
import { POST } from "@starfederation/datastar/dist/plugins/official/backend/actions/post";
import { PUT } from "@starfederation/datastar/dist/plugins/official/backend/actions/put";
import { Indicator } from "@starfederation/datastar/dist/plugins/official/backend/attributes/indicator";
import { ExecuteScript } from "@starfederation/datastar/dist/plugins/official/backend/watchers/executeScript";
import { MergeFragments } from "@starfederation/datastar/dist/plugins/official/backend/watchers/mergeFragments";
import { MergeSignals } from "@starfederation/datastar/dist/plugins/official/backend/watchers/mergeSignals";
import { RemoveFragments } from "@starfederation/datastar/dist/plugins/official/backend/watchers/removeFragments";
import { RemoveSignals } from "@starfederation/datastar/dist/plugins/official/backend/watchers/removeSignals";
import { Clipboard } from "@starfederation/datastar/dist/plugins/official/browser/actions/clipboard";
import { Show } from "@starfederation/datastar/dist/plugins/official/browser/attributes/show";
import { ViewTransition } from "@starfederation/datastar/dist/plugins/official/browser/attributes/viewTransition";
import { Attr } from "@starfederation/datastar/dist/plugins/official/dom/attributes/attr";
import { Bind } from "@starfederation/datastar/dist/plugins/official/dom/attributes/bind";
import { Class } from "@starfederation/datastar/dist/plugins/official/dom/attributes/class";
import { On } from "@starfederation/datastar/dist/plugins/official/dom/attributes/on";
import { Ref } from "@starfederation/datastar/dist/plugins/official/dom/attributes/ref";
import { Text } from "@starfederation/datastar/dist/plugins/official/dom/attributes/text";
import "instant.page/instantpage";

datastar.load(
  GET,
  DELETE,
  PATCH,
  POST,
  PUT,
  Indicator,
  ExecuteScript,
  MergeFragments,
  MergeSignals,
  RemoveFragments,
  RemoveSignals,
  Clipboard,
  Show,
  ViewTransition,
  Attr,
  Bind,
  Class,
  On,
  Ref,
  Text
);

// Initialize Datastar
datastar.apply();

// store image cache in browser local storage
function fetchAndCacheImage(
  img: HTMLImageElement,
  src: string,
  fallbackSrc?: string
): void {
  fetch(src)
    .then((response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      return response.blob();
    })
    .then((blob) => {
      const reader = new FileReader();
      reader.onloadend = () => {
        const dataURL = reader.result as string;
        localStorage.setItem(src, dataURL); // Cache the image
        img.src = dataURL; // Set the image source
      };
      reader.readAsDataURL(blob);
    })
    .catch((error) => {
      console.error("Failed to fetch image:", error);

      // Display a fallback image
      if (fallbackSrc) {
        img.src = fallbackSrc;
      }
    });
}

// cache images if there is a data-src property
document.addEventListener("DOMContentLoaded", () => {
  const images = document.querySelectorAll(
    "img[data-src]"
  ) as NodeListOf<HTMLImageElement>;
  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement;
        const src = img.getAttribute("data-src") as string;
        const fallbackSrc = img.getAttribute("data-fallback-src") as
          | string
          | undefined;
        // Check if the image is already cached
        const cachedImage = localStorage.getItem(src);
        if (cachedImage) {
          img.src = cachedImage;
        } else {
          // Fetch and cache the image
          fetchAndCacheImage(img, src, fallbackSrc);
        }
        observer.unobserve(img); // Stop observing once the image is loaded
      }
    });
  });

  images.forEach((img) => observer.observe(img));
});
