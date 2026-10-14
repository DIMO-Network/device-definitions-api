import { readFile } from 'node:fs/promises';
import { mapAttributes } from '../../../../../definitions-worker/scripts/extract/vocabulary.mjs';
const vocab = JSON.parse(await readFile(new URL('./device-type-vehicle.json', import.meta.url),'utf8'));
const cases = [
  {fuel_type:'Petrol'}, {fuel_type:'Hybrid'}, {fuel_type:'Flexible Fuel Vehicle (FFV)'},
  {driven_wheels:'4x2'}, {driven_wheels:'AWD/All-Wheel Drive'}, {driven_wheels:'4WD/4-Wheel Drive/4x4'},
  {epa_class:'6L'}, {epa_class:'Euro 6d-Temp'}, {epa_class:'Calss III'},
  {vehicle_type:'Crossover Utility Vehicle (CUV)'}, {vehicle_type:'Incomplete Vehicle'},
  {wheelbase:'111.2 in'}, {wheelbase:'abc'},
  {number_of_doors:'4'}, {number_of_doors:'4.5'}, {number_of_doors:'99'},
  {fuel_tank_capacity_gal:'15.800000'}, {mpg:'28'},
  {generation:'6'}, {fuel_type:''}, {fuel_type:'<nil>'}, {powertrain_type:'ICE'},
];
const out = cases.map((c) => {
  const { mapped, dropped } = mapAttributes(c, vocab);
  return { in: c, mapped, dropped: dropped.map(d => ({name: d.name, reason: d.reason})) };
});
console.log(JSON.stringify(out, null, 1));
