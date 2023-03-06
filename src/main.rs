extern crate ffmpeg_next as ffmpeg;

use ffmpeg::format::{input, Pixel};
use ffmpeg::media::Type;
use ffmpeg::software::scaling::{context::Context, flag::Flags};
use ffmpeg::util::frame::video::Video;
use std::env;
use std::error::Error;
use std::fs::File;
use std::io::prelude::*;

fn main() -> Result<(), Box<dyn Error>> {
    ffmpeg::init().unwrap();

    if let Ok(mut ictx) = input(&env::args().nth(1).expect("Cannot open file.")) {
        let input = ictx
            .streams()
            .best(Type::Video)
            .ok_or(ffmpeg::Error::StreamNotFound)?;
        let video_stream_index = input.index();

        let context_decoder = ffmpeg::codec::context::Context::from_parameters(input.parameters())?;
        let mut decoder = context_decoder.decoder().video()?;

        let mut scaler = Context::get(
            decoder.format(),
            decoder.width(),
            decoder.height(),
            Pixel::RGB24,
            decoder.width(),
            decoder.height(),
            Flags::BILINEAR,
        )?;

        let mut frame_index = 0;
        let mut column_colours: Vec<Vec<Colour>> = Vec::new();

        let mut receive_and_process_decoded_frames =
            |decoder: &mut ffmpeg::decoder::Video| -> Result<(), ffmpeg::Error> {
                let mut decoded = Video::empty();
                while decoder.receive_frame(&mut decoded).is_ok() {
                    let mut rgb_frame = Video::empty();
                    scaler.run(&decoded, &mut rgb_frame)?;

                    println!("frame{}", frame_index);

                    column_colours.push(get_avg_colours(&rgb_frame));

                    frame_index +=1 ;
                }
                Ok(())
            };

        for (stream, packet) in ictx.packets() {
            if stream.index() == video_stream_index {
                decoder.send_packet(&packet)?;
                receive_and_process_decoded_frames(&mut decoder)?;
            }
        }

        decoder.send_eof()?;
        receive_and_process_decoded_frames(&mut decoder)?;

        println!("Saving frame");

        let width = column_colours.len();
        let height = column_colours[0].len();

        let final_frame = get_final_frame(&column_colours);
        save_frame(&final_frame, width as u32, height as u32)?;
    }

    Ok(())
}

struct Colour {
    r: u8,
    g: u8,
    b: u8,
}

fn get_avg_colours(frame: &Video) -> Vec<Colour> {
    let data = frame.data(0);

    let mut colours: Vec<Colour> = Vec::new();

    for i in (0..data.len()).step_by(3) {
        let colour = Colour {
            r: data[i],
            g: data[i + 1],
            b: data[i + 2],
        };

        colours.push(colour);
    }

    return (0..frame.height() as usize)
        .into_iter()
        .map(|i| &colours[i..frame.width() as usize])
        .map(|colours| {
            let r = (colours.iter().map(|c| c.r as usize).sum::<usize>() / colours.len()) as u8;
            let g = (colours.iter().map(|c| c.g as usize).sum::<usize>() / colours.len()) as u8;
            let b = (colours.iter().map(|c| c.b as usize).sum::<usize>() / colours.len()) as u8;

            return Colour { r, g, b };
        })
        .collect();
}

fn get_final_frame(column_colours: &Vec<Vec<Colour>>) -> Vec<u8> {
    let width = column_colours.len();
    let height = column_colours[0].len();

    let mut frame: Vec<u8> = Vec::new();

    for i in 0..height {
        for j in 0..width {
            let col = &column_colours[j][i];

            frame.push(col.r);
            frame.push(col.g);
            frame.push(col.b);
        }
    }

    return frame;
}

fn save_frame(frame: &Vec<u8>, width: u32, height: u32) -> std::result::Result<(), std::io::Error> {
    let mut file = File::create("out.ppm")?;
    file.write_all(format!("P6\n{} {}\n255\n", width, height).as_bytes())?;
    file.write_all(frame)?;
    Ok(())
}
